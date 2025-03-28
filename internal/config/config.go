package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func InitConfig() error {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Error().Msg("Error loading .env file")
		return err
	}

	return nil
}
