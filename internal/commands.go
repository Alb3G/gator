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

type Command struct {
	Name string
	Args []string
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
		log.Fatal(err)
	}

	s.Config.SetUser(userName)

	return nil
}

func RegisterHandler(s *conf.State, c Command) error {
	// Add a util function in the future to validate correct userNames
	if len(c.Args) != 2 {
		log.Fatal("no user name provided")
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
		log.Fatal(err)
	}

	if userFromDb.UserName != "" {
		log.Fatal("user_name already exists in db")
	}

	user, err := s.Queries.CreateUser(ctx, dbArgs)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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
	feedURL := "https://www.wagslane.dev/index.xml"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rssStruct, err := rss.FetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}

	fmt.Println(rssStruct)

	return nil
}

func AddFeed(s *conf.State, c Command) error {
	if len(c.Args) < 3 {
		log.Fatal("missing required args feed_name or url")
	}
	name := c.Args[1]
	url := c.Args[2]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.Queries.GetUserByName(ctx, s.Config.CurrentUserName)
	if err != nil {
		log.Fatal(err)
	}

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
		log.Fatal(err)
	}

	fmt.Println(feed)

	return nil
}
