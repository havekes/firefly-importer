package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	FireflyURL   string
	FireflyToken string
	VisionAPIURL string
	VisionAPIKey string
	VisionModel  string
	Port         string
}

func LoadConfig() *Config {
	// Attempt to load from .env file, ignore error if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading configuration from environment variables")
	}

	config := &Config{
		FireflyURL:   os.Getenv("FIREFLY_URL"),
		FireflyToken: os.Getenv("FIREFLY_TOKEN"),
		VisionAPIURL: os.Getenv("VISION_API_URL"),
		VisionAPIKey: os.Getenv("VISION_API_KEY"),
		VisionModel:  os.Getenv("VISION_API_MODEL"),
		Port:         os.Getenv("PORT"),
	}

	// Set defaults
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.FireflyURL == "" {
		config.FireflyURL = "https://firefly.havek.es/api/v1"
	}
	if config.VisionModel == "" {
		config.VisionModel = "gpt-4-vision-preview"
	}

	return config
}
