package forestryroads

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

// implementedMethods is a list of the implemented HTTP methods for the status endpoint.
var implementedMethods = []string{http.MethodGet}

// Handler
// Currently only supports GET requests.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	// Switch on the HTTP request method
	switch r.Method {
	case http.MethodGet:
		handleForestryRoadGet(w, r)

	default:
		// If the method is not implemented, return an error with the allowed methods
		http.Error(
			w, fmt.Sprintf(
				"REST Method '%s' not supported. Currently only '%v' are supported.", r.Method,
				implementedMethods,
			), http.StatusNotImplemented,
		)
		return
	}
}

// handleForestryRoadGet handles GET requests to the forestry road endpoint.
func handleForestryRoadGet(w http.ResponseWriter, r *http.Request) {
	// Pseudo code
	// 1. Mirror the request to the remote server
	// 2. Get the response from the remote server
	// 3. Parse the response
	// 4. Calculate trafficality
	// 5. Return the response, with the calculated trafficality as a rgb value in the geojson response

	// Get time parameter from url
	time := r.URL.Query().Get("time")
	// Split ISO string to get date. ex: 2021-03-01T00:00:00Z -> 2021-03-01
	// Gets put into struct later
	date := strings.Split(time, "T")[0]

	// Mirror request to https://wms.geonorge.no/skwms1/wms.traktorveg_skogsbilveger
	proxyReq, err := http.NewRequest(r.Method, "https://wms.geonorge.no/skwms1/wms.traktorveg_skogsbilveger?"+r.URL.RawQuery, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		log.Println("Error creating request: ", err)
		return
	}

	// Do request
	proxyResp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to fetch data from WMS server", http.StatusBadGateway)
		log.Println("Error fetching data from WMS server: ", err)
		return
	}

	// Decode into struct
	var wfsResponse WFSResponse
	err = json.NewDecoder(proxyResp.Body).Decode(&wfsResponse)
	if err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		log.Println("Error decoding response: ", err)
		return
	}

	// Randomize color for testing, update date
	for i, _ := range wfsResponse.Features {
		if wfsResponse.Date == "" {
			wfsResponse.Date = date
		}

		if wfsResponse.Features[i].Properties.Farge == nil {
			wfsResponse.Features[i].Properties.Farge = make([]int, 3)
		}

		wfsResponse.Features[i].Properties.Farge[0] = rand.Intn(256)
		wfsResponse.Features[i].Properties.Farge[1] = rand.Intn(256)
		wfsResponse.Features[i].Properties.Farge[2] = rand.Intn(256)
	}

	// Encode response
	err = json.NewEncoder(w).Encode(wfsResponse)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Println("Error encoding response: ", err)
		return
	}
}
