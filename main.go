package main

import (
	"fmt"

	"github.com/Alb3G/gator/internal/config"
)

func main() {
	c := config.Read()
	fmt.Printf("User name before change: %v\n", c.CurrentUserName)
	c.SetUser("Alb3G")
	fmt.Printf("User name after change: %v\n", c.CurrentUserName)
	fmt.Println(c)
}
