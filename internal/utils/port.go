package utils

import (
	"github.com/rs/zerolog/log"
	"os"
)

// _defaultPort is the default port for the server
const _defaultPort = "8080"

// GetPort gets the port from the environment variable, or uses the default port; 8080.
func GetPort() string {
	// Get the PORT environment variable
	port := os.Getenv("PORT")

	// Use default Port variable if not provided
	if port == "" {
		log.Warn().Msg("$PORT has not been set. Default: " + _defaultPort)
		port = _defaultPort
	}

	return port
}
