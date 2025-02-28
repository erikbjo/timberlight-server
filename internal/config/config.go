package config

import (
	"github.com/joho/godotenv"
	"log"
)

func InitConfig() error {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
		return err
	}

	return nil
}
