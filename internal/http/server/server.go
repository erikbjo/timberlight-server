package server

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/http/handlers"
	"skogkursbachelor/server/internal/utils"
)

// Start starts the server on the port specified in the environment variable PORT.
// If PORT is not set, the default port 8080 is used.
func Start() {
	// Get the port from the environment variable, or use the default port
	port := utils.GetPort()

	// Get list of proxy endpoints
	proxies, err := utils.LoadProxiesFromFile()
	if err != nil {
		log.Fatal().Msg("Error loading proxies: " + err.Error())
	}

	// Using mux to handle /'s and parameters
	mux := http.NewServeMux()

	// Set up handler endpoints, with and without trailing slash
	// Proxies
	for path, remoteAddr := range proxies {
		log.Info().Msg(path + "->" + remoteAddr)
		p := &handlers.Proxy{RemoteAddr: remoteAddr}
		mux.HandleFunc(constants.ProxyPath+path, p.ProxyHandler)
	}

	// Base layer
	mux.HandleFunc(constants.BaseLayerPath+"/{type}/{abc}/{z}/{x}/{y}", handlers.BaseLayerHandler)
	mux.HandleFunc(constants.BaseLayerPath+"/{type}/{abc}/{z}/{x}", handlers.BaseLayerHandler)

	// Forestry roads
	mux.HandleFunc(constants.ForestryRoadsPath, handlers.ForestryRoadsHandler)

	// Start server
	log.Info().Msg("Starting server on port " + port + " ...")
	log.Fatal().Msg(http.ListenAndServe(":"+port, mux).Error())
}
