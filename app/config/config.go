package config

import (
	flags "github.com/jessevdk/go-flags"
	"log"
	"os"
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

	err = SetDiscoveryAsLocalhostEnvVar()
	if err != nil {
		return cfg, err
	}

	return cfg, err
}

func SetDiscoveryAsLocalhostEnvVar() error {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}
	return err
}
