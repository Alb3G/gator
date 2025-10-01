package internal

import (
	"errors"

	conf "github.com/Alb3G/gator/internal/config"
)

type Commands struct {
}

type Command struct {
	Name string
	Args []string
}

func loginHandler(s *conf.State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("missing username argument")
	}

	s.Config.SetUser(cmd.Args[0])

	return nil
}
