// Package main is the entry point for the application, it starts the server.
package main

import (
	"os"
	"skogkursbachelor/server/internal/config"
	"skogkursbachelor/server/internal/http/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Start the server
func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if err := config.InitConfig(); err != nil {
		log.Fatal().Msgf("Error loading configuration: %s", err)
	}

	loggerLevel := os.Getenv("LOGGER_LEVEL")
	if loggerLevel == "" {
		loggerLevel = "info"
	}
	lvl, err := zerolog.ParseLevel(loggerLevel)
	if err != nil {
		log.Fatal().Msgf("Error parsing log level: %s", err)
	}
	zerolog.SetGlobalLevel(lvl)
	log.Info().Msgf("Logger level set to %s", lvl)

	log.Info().Msg("Configuration loaded successfully")

	server.Start()
}
