package internal

import (
	"context"
	"errors"
	"time"

	conf "github.com/Alb3G/gator/internal/config"
	"github.com/Alb3G/gator/internal/database"
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
		return err
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

	s.Config.SetUser(cmd.Args[1])

	return nil
}

func RegisterHandler(s *conf.State, c Command) error {
	uuid := uuid.New()
	args := database.CreateUserParams{
		ID:        uuid,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserName:  c.Args[1],
	}
	// waiting to see what we do with the user
	_, err := s.Queries.CreateUser(context.Background(), args)
	if err != nil {
		return err
	}

	return nil
}
