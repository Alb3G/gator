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
	utils "github.com/Alb3G/gator/internal/utils"
	"github.com/google/uuid"
)

type Commands struct {
	AvailableCommands map[string]func(*conf.State, Command) error
}

// This method runs a given command with the provided state if it exists.
func (c *Commands) Run(state *conf.State, cmd Command) error {
	f := c.AvailableCommands[cmd.Name]

	err := f(state, cmd)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// This method registers a new handler function for a command name.
func (c *Commands) Register(name string, f func(*conf.State, Command) error) error {
	c.AvailableCommands[name] = f

	return nil
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

	exists, err := s.Queries.GetUserByName(ctx, userName)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}

	if exists != "" {
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
