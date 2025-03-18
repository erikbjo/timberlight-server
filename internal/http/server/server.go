package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/http/handlers/forestryroads"
	"skogkursbachelor/server/internal/http/handlers/proxy"
	"skogkursbachelor/server/internal/utils"
)

func loadConfig(file string) (map[string]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config map[string]string
	err = json.Unmarshal(data, &config)
	return config, err
}

// Start
/*
Start the server on the port specified in the environment variable PORT. If PORT is not set, the default port 8080 is used.
*/
func Start() {
	// Get the port from the environment variable, or use the default port
	port := utils.GetPort()

	// Get list of proxy endpoints
	proxies, err := loadConfig("proxy.json")
	if err != nil {
		log.Fatal("Error loading proxies: ", err)
	}

	// Using mux to handle /'s and parameters
	mux := http.NewServeMux()

	// Set up handler endpoints, with and without trailing slash
	// Proxies
	for path, remoteAddr := range proxies {
		log.Println(path, "->", remoteAddr)
		p := &proxy.Proxy{RemoteAddr: remoteAddr}
		mux.HandleFunc(constants.ProxyPath+path, p.ProxyHandler)
	}

	// Forestry roads
	mux.HandleFunc(constants.ForestryRoadsPath, forestryroads.Handler)

	// Start server
	log.Println("Starting server on port " + port + " ...")
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
