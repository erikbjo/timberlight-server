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

	// Test frost
	//testFeature := forestryroads.WFSFeature{
	//	Type: "Feature",
	//	Properties: struct {
	//		Kommunenummer      string `json:"kommunenummer"`
	//		Vegkategori        string `json:"vegkategori"`
	//		Vegfase            string `json:"vegfase"`
	//		Vegnummer          string `json:"vegnummer"`
	//		Strekningnummer    string `json:"strekningnummer"`
	//		Delstrekningnummer string `json:"delstrekningnummer"`
	//		Frameter           string `json:"frameter"`
	//		Tilmeter           string `json:"tilmeter"`
	//		Farge              []int  `json:"farge"`
	//	}{
	//		Kommunenummer:      "1234",
	//		Vegkategori:        "V",
	//		Vegfase:            "1",
	//		Vegnummer:          "1",
	//		Strekningnummer:    "1",
	//		Delstrekningnummer: "1",
	//		Frameter:           "1",
	//		Tilmeter:           "1",
	//		Farge:              []int{255, 255, 0},
	//	},
	//	Geometry: struct {
	//		Type        string      `json:"type"`
	//		Coordinates [][]float64 `json:"coordinates"`
	//	}{
	//		Type: "LineString",
	//		Coordinates: [][]float64{
	//			{1201320.00765347, 8535377.37942303},
	//			{1201319.16169355, 8535358.78993844},
	//			{1201318.00905091, 8535335.45018047},
	//			{1201316.12991158, 8535313.91938757},
	//		},
	//	},
	//}
	//
	//log.Println(forestryroads.GetIsGroundFrozenAlongFeature(testFeature, "2025-03-18"))

	//coords := []float64{1186244.298553594, 8579340.020600447}
	//log.Println(forestryroads.GetIsGroundFrozenAlongFeature(coords, "2025-03-17"))

	// Start server
	log.Println("Starting server on port " + port + " ...")
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
