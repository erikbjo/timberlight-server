package server

import (
	"log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/http/proxy"
	"skogkursbachelor/server/internal/utils"
)

// Start
/*
Start the server on the port specified in the environment variable PORT. If PORT is not set, the default port 8080 is used.
*/
func Start() {
	// Get the port from the environment variable, or use the default port
	port := utils.GetPort()

	// Using mux to handle /'s and parameters
	mux := http.NewServeMux()

	// Set up handler endpoints, with and without trailing slash
	// Proxy
	testProxy := proxy.Proxy{
		RemoteAddr: "http://example.com/",
	}

	mux.HandleFunc(constants.ProxyPath, testProxy.ProxyHandler)

	// Start server
	log.Println("Starting server on port " + port + " ...")
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
