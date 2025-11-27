package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config структура
type Config struct {
	HTTPServer   `yaml:"http_server"`
	DBConnection `yaml:"db_path"`
}

type DBConnection struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port" env:"PORT" env-required:"true"`
	Username string `yaml:"username" env:"DB_USER" env-required:"true"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"localhost:8081"`
}

func MustLoad() *Config {
	const op = "config.config.MustLoad"
	// Load .env file if it exists (optional for Docker environments)
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
