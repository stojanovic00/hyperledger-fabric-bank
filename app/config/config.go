package config

import (
	flags "github.com/jessevdk/go-flags"
)

type Config struct {
	Host      string `long:"host" env:"SDK_APP_HOST"`
	Port      string `long:"port" env:"SDK_APP_PORT"`
	JWTSecret string `long:"secret" env:"JWT_SECRET"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	parser := flags.NewParser(&cfg, flags.Default)
	_, err := parser.Parse()
	return cfg, err
}
