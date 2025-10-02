package main

import (
	"fmt"
	"os"

	"github.com/Alb3G/gator/internal"
	"github.com/Alb3G/gator/internal/config"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("No enough arguments provided")
		os.Exit(1)
	}

	c := config.Read()

	s := config.State{Config: c}

	cmds := internal.Commands{
		AvailableCommands: make(map[string]func(*config.State, internal.Command) error),
	}

	cmds.Register("login", internal.LoginHandler)

	cmd := internal.Command{
		Name: args[1],
		Args: args[1:],
	}

	err := cmds.Run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
