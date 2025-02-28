// Package main is the entry point for the application, it starts the server.
package main

import (
	"log"
	"skogkursbachelor/server/internal/config"
	"skogkursbachelor/server/internal/http/server"
)

func init() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}
	log.Println("Configuration loaded successfully")
}

// Start the server
func main() {
	server.Start()
}
