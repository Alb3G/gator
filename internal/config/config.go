package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Alb3G/gator/internal/database"
)

const CONFIG_FILE = ".gatorconfig.json"

type State struct {
	Config  *Config
	Queries *database.Queries
}

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(userName string) {
	c.CurrentUserName = userName

	c.write()

	fmt.Println("New User has been set.")
}

func (c *Config) write() {
	path, err := getConfigFilePath()
	if err != nil {
		log.Fatal(err)
	}

	newConfBytes, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(path, newConfBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func Read() *Config {
	filePath, err := getConfigFilePath()
	if err != nil {
		log.Fatal(err)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var c Config
	err = json.NewDecoder(bytes.NewBuffer(b)).Decode(&c)
	if err != nil {
		log.Fatal(err)
	}

	return &c
}

func getConfigFilePath() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("%v/%v", userHomeDir, CONFIG_FILE)

	return path, nil
}
