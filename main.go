package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Alb3G/gator/internal"
	"github.com/Alb3G/gator/internal/config"
	"github.com/Alb3G/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("No enough arguments provided")
		os.Exit(1)
	}

	c := config.Read()

	db, err := sql.Open("postgres", c.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	queries := database.New(db)

	s := config.State{Config: c, Queries: queries}

	cmds := internal.Commands{
		AvailableCommands: make(map[string]func(*config.State, internal.Command) error),
	}

	cmds.Register("login", internal.LoginHandler)
	cmds.Register("register", internal.RegisterHandler)
	cmds.Register("reset", internal.ResetHandler)
	cmds.Register("users", internal.Users)
	cmds.Register("agg", internal.Agg)
	cmds.Register("addfeed", internal.AddFeed)
	cmds.Register("feeds", internal.FeedsHandler)

	cmd := internal.Command{
		Name: args[1],
		Args: args[1:],
	}

	err = cmds.Run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
