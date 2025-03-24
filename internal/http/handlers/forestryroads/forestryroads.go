package forestryroads

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/utils"
	"strings"
	"sync"
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

func roundToNearest500(n int) int {
	base := (n / 1000) * 1000
	return base + 500
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
	if time == "" {
		http.Error(w, "Missing time URL parameter", http.StatusBadRequest)
		return
	}

	// Split ISO string to get date. ex: 2021-03-01T00:00:00Z -> 2021-03-01
	// Gets put into struct later
	date := strings.Split(time, "T")[0]

	// Mirror request to https://wms.geonorge.no/skwms1/wms.traktorveg_skogsbilveger
	proxyReq, err := http.NewRequest(
		r.Method,
		constants.ForestryRoadsWFS+"?"+r.URL.RawQuery,
		r.Body,
	)
	if err != nil {
		http.Error(w, "Failed to create internal request", http.StatusInternalServerError)
		log.Println("Error creating request to GeoNorge for forestry roads: ", err)
		return
	}

	// Do request
	proxyResp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to fetch data from external WMS server", http.StatusBadGateway)
		log.Println("Error fetching data from GeoNorge WMS server: ", err)
		return
	}

	// Decode into struct
	var wfsResponse WFSResponse
	err = json.NewDecoder(proxyResp.Body).Decode(&wfsResponse)
	if err != nil {
		http.Error(w, "Failed to decode external response", http.StatusInternalServerError)
		log.Println("Error decoding response from GeoNorge WMS server: ", err)
		return
	}

	// Group the features by EPSG25833 coordinates, each with a cluster at coordinates: xxx500, yyy500
	// This is a center point of a 1000x1000 meter square, and the center of SeNorge grid cells
	// This is done to reduce the number of requests to the frost API, as the frost API is slow and cannot
	// handle requests with multiple coordinates in the same grid cell
	// This is done by creating a map with the key being the coordinates, and the value being a list of features
	// with the same coordinates
	// This is done by iterating over the features, and for each feature, we get the coordinates, and round them
	// to the nearest 500, and add the feature to the list of features at that key

	// Create a map to store the features
	featureMap := make(map[string][]WFSFeature)

	// Iterate over the features
	for _, feature := range wfsResponse.Features {
		// Get the middle of the road (ish), a feature is small enough to only have one coordinate
		middleIndex := len(feature.Geometry.Coordinates) / 2
		coordinates := feature.Geometry.Coordinates[middleIndex]

		// Transform the coordinates to EPSG:25833
		newX, newY, err := utils.TransformCoordinates(coordinates, 3857, 25833)
		if err != nil {
			http.Error(w, "Failed to transform coordinates", http.StatusInternalServerError)
			log.Println("Error transforming coordinates: ", err)
			return
		}

		roundedX := roundToNearest500(newX)
		roundedY := roundToNearest500(newY)

		// Round the coordinates to the nearest 500
		roundedCoordinates := fmt.Sprintf("%d,%d", roundedX, roundedY)

		// Update struct with new coordinates
		feature.MiddleOfRoad25833 = []int{roundedX, roundedY}

		// Add the feature to the list of features at the key
		featureMap[roundedCoordinates] = append(featureMap[roundedCoordinates], feature)
	}

	var wg sync.WaitGroup
	var transcribedFeatures []WFSFeature
	var featuresMutex sync.Mutex

	semaphore := make(chan struct{}, 500)

	// Iterate over the features
	for _, features := range featureMap {
		semaphore <- struct{}{} // Reserve a slot
		wg.Add(1)

		go func(features []WFSFeature) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the slot

			if wfsResponse.Date == "" {
				wfsResponse.Date = date
			}

			// Get first feature, as all features have the same coordinates
			feature := features[0]

			isFrozen, err := GetIsGroundFrozen(feature.MiddleOfRoad25833, date)
			if err != nil {
				http.Error(w, "Failed to get frost data", http.StatusInternalServerError)
				log.Println("Error getting frost data: ", err)
				return
			}

			// Loop over the features and set the color based on the frost data
			for i := range features {
				if features[i].Properties.Farge == nil {
					features[i].Properties.Farge = make([]int, 3)
				}

				// If the ground is frozen, set the color to green
				if isFrozen {
					features[i].Properties.Farge[0] = 0
					features[i].Properties.Farge[1] = 255
					features[i].Properties.Farge[2] = 0
				} else {
					// If the ground is not frozen, set the color to red
					features[i].Properties.Farge[0] = 255
					features[i].Properties.Farge[1] = 0
					features[i].Properties.Farge[2] = 0
				}
			}

			featuresMutex.Lock()
			transcribedFeatures = append(transcribedFeatures, features...)
			featuresMutex.Unlock()
		}(features)
	}

	wg.Wait()
	// Replace the features with the transcribed features
	wfsResponse.Features = transcribedFeatures

	// Encode response
	err = json.NewEncoder(w).Encode(wfsResponse)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Println("Error encoding final response: ", err)
		return
	}
}
