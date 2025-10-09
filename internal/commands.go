package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	conf "github.com/Alb3G/gator/internal/config"
	"github.com/Alb3G/gator/internal/database"
	rss "github.com/Alb3G/gator/internal/rss"
	utils "github.com/Alb3G/gator/internal/utils"
	uuid "github.com/google/uuid"
)

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	AvailableCommands map[string]func(*conf.State, Command) error
}

// This method runs a given command with the provided state if it exists.
func (c *Commands) Run(state *conf.State, cmd Command) error {
	f, ok := c.AvailableCommands[cmd.Name]
	if !ok {
		return errors.New("command not found")
	}

	return f(state, cmd)
}

// This method registers a new handler function for a command name.
func (c *Commands) Register(name string, f func(*conf.State, Command) error) {
	c.AvailableCommands[name] = f
}

func MiddlewareLoggedIn(handler func(s *conf.State, c Command, user database.User) error) func(s *conf.State, c Command) error {
	return func(s *conf.State, c Command) error {
		user, err := s.Queries.GetUserByName(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			return err
		}

		err = handler(s, c, user)
		if err != nil {
			return err
		}

		return nil
	}
}

func LoginHandler(s *conf.State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return errors.New("missing username argument")
	}

	userName := cmd.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.Queries.GetUserByName(ctx, userName)
	if err != nil {
		return err
	}

	s.Config.SetUser(userName)

	return nil
}

func RegisterHandler(s *conf.State, c Command) error {
	// Add a util function in the future to validate correct userNames
	if len(c.Args) != 2 {
		return errors.New("no user name provided")
	}

	userName := c.Args[1]
	uuid := uuid.New()
	dbArgs := database.CreateUserParams{
		ID:        uuid,
		CreatedAt: utils.Now(),
		UpdatedAt: utils.Now(),
		UserName:  userName,
	}

	// Generate context with Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userFromDb, err := s.Queries.GetUserByName(ctx, userName)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if userFromDb.UserName != "" {
		return errors.New("user_name already exists in db")
	}

	user, err := s.Queries.CreateUser(ctx, dbArgs)
	if err != nil {
		return err
	}

	fmt.Println("User created successfully")
	fmt.Println(user)

	s.Config.SetUser(userName)

	return nil
}

func ResetHandler(s *conf.State, c Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Queries.Reset(ctx)
	if err != nil {
		return err
	}

	return nil
}

func Users(s *conf.State, c Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := s.Queries.GetUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.UserName == s.Config.CurrentUserName {
			fmt.Printf("* %v (current)", s.Config.CurrentUserName)
		} else {
			fmt.Print("* ")
			fmt.Println(user.UserName)
		}

	}

	return nil
}

func Agg(s *conf.State, c Command) error {
	if len(c.Args) < 2 {
		return errors.New("missing time_between_reqs arg")
	}
	time_between_reqs, err := time.ParseDuration(c.Args[1])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %v\n", time_between_reqs)

	ticker := time.NewTicker(time_between_reqs)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func AddFeed(s *conf.State, c Command, user database.User) error {
	if len(c.Args) < 3 {
		return errors.New("missing required args feed_name or url")
	}
	name := c.Args[1]
	url := c.Args[2]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedArgs := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.Queries.CreateFeed(ctx, feedArgs)
	if err != nil {
		return err
	}

	feed_follow_Args := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.Queries.CreateFeedFollow(ctx, feed_follow_Args)
	if err != nil {
		return err
	}

	fmt.Println(feed)

	return nil
}

func FeedsHandler(s *conf.State, c Command, user database.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feeds, err := s.Queries.GetFeeds(ctx)
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println(user.UserName)
		fmt.Println(feed)
	}

	return nil
}

func scrapeFeeds(s *conf.State) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lastFeedFetched, err := s.Queries.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	feedFetchedParams := database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		UpdatedAt: time.Now().UTC(),
		ID:        lastFeedFetched.ID,
	}

	err = s.Queries.MarkFeedFetched(ctx, feedFetchedParams)
	if err != nil {
		return err
	}

	rssStruct, err := rss.FetchFeed(ctx, lastFeedFetched.Url)
	if err != nil {
		return err
	}

	for _, item := range rssStruct.Channel.Item {
		pubDate, err := utils.ParsePublishedDate(item.PubDate)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			return err
		}

		postParams := database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			PublishedAt: pubDate,
			FeedID:      lastFeedFetched.ID,
		}
		_, err = s.Queries.CreatePost(ctx, postParams)
		if err != nil {
			log.Printf("Error creating post: %v", err)
			return err
		}
	}

	return nil
}

func Follow(s *conf.State, c Command, user database.User) error {
	if len(c.Args) < 2 {
		return errors.New("missing url arg")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feed, err := s.Queries.GetFeedByURL(ctx, c.Args[1])
	if err != nil {
		return err
	}

	feed_follow_Args := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	inserted_feed_follow, err := s.Queries.CreateFeedFollow(ctx, feed_follow_Args)
	if err != nil {
		return err
	}

	fmt.Printf("%v followed: %v", user.UserName, inserted_feed_follow.FeedName)

	return nil
}

func Following(s *conf.State, c Command, user database.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedFollowsByUser, err := s.Queries.GetFeedFollowsByUser(ctx, user.ID)
	if err != nil {
		return err
	}

	for _, feed_follow := range feedFollowsByUser {
		fmt.Println(feed_follow.FeedName)
	}

	return nil
}

func Unfollow(s *conf.State, c Command, user database.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feed, err := s.Queries.GetFeedByURL(ctx, c.Args[1])
	if err != nil {
		return err
	}

	deleteParams := database.DeleteFeedFollowParams{UserID: user.ID, FeedID: feed.ID}

	err = s.Queries.DeleteFeedFollow(ctx, deleteParams)
	if err != nil {
		return err
	}

	return nil
}

func Browse(s *conf.State, c Command, user database.User) error {
	limit := utils.ParseLimit(c.Args, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postsParams := database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  limit,
	}
	posts, err := s.Queries.GetPostsByUser(ctx, postsParams)
	if err != nil {
		log.Printf("Error while getting posts from db: %v", err)
		return err
	}

	if len(posts) == 0 {
		log.Println("No posts found")
		return nil
	}

	for _, post := range posts {
		fmt.Println(post)
	}

	return nil
}
