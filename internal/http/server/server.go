package server

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/http/handlers/forestryroads"
	"skogkursbachelor/server/internal/http/handlers/proxy"
	"skogkursbachelor/server/internal/utils"
)

// Start
/*
Start the server on the port specified in the environment variable PORT. If PORT is not set, the default port 8080 is used.
*/
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
		p := &proxy.Proxy{RemoteAddr: remoteAddr}
		mux.HandleFunc(constants.ProxyPath+path, p.ProxyHandler)
	}

	// Forestry roads
	mux.HandleFunc(constants.ForestryRoadsPath, forestryroads.Handler)

	// Start server
	log.Info().Msg("Starting server on port " + port + " ...")
	log.Fatal().Msg(http.ListenAndServe(":"+port, mux).Error())
}
