package config

type Config struct {
	DbUrl           string
	CurrentUserName string
}

func Read() *Config {
	return &Config{}
}
