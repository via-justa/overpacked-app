package config

import (
	"fmt"
	"os"
)

const (
	defaultServerAddr = "0.0.0.0:8000"
	defaultImagesDir  = "./data/images"
)

type Config struct {
	DatabaseURL string
	ServerAddr  string
	AppUsername string
	AppPassword string
	JWTSecret   string
	ImagesDir   string
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = defaultServerAddr
	}

	imagesDir := os.Getenv("IMAGES_DIR")
	if imagesDir == "" {
		imagesDir = defaultImagesDir
	}

	return &Config{
		DatabaseURL: databaseURL,
		ServerAddr:  serverAddr,
		AppUsername: os.Getenv("APP_USERNAME"),
		AppPassword: os.Getenv("APP_PASSWORD"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		ImagesDir:   imagesDir,
	}, nil
}
