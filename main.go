package main

import (
	"fmt"

	"github.com/Alb3G/gator/internal/config"
)

func main() {
	c := config.Read()
	c.SetUser("Alb3G")
	fmt.Println(c)
}
