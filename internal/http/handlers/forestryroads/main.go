package forestryroads

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"strings"
	"time"
)

// implementedMethods is a list of the implemented HTTP methods for the status endpoint.
var implementedMethods = []string{http.MethodGet}

// index is a spatial index for the forestry roads
var index *SpatialIndex

func init() {
	shapefiles := []string{
		"data/Losmasse/LosmasseFlate_20240621",
		"data/Losmasse/LosmasseFlate_20240622",
		"data/Losmasse/LosmasseGrense_20240621",
	}

	// Build spatial index
	index = ReadShapeFilesAndBuildIndex(shapefiles)
	log.Println("Index built successfully!")

	log.Println("Forestry roads handler initialized")
}

// Handler handles requests to the forestry road endpoint.
// Currently only GET requests are supported.
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
	startTotal := time.Now()
	// Pseudo code
	// 1. Mirror the request to the remote server
	// 2. Get the response from the remote server
	// 3. Parse the response
	// 4. Calculate trafficality
	// 5. Return the response, with the calculated trafficality as a rgb value in the geojson response

	// Get timeDate parameter from url
	timeDate := r.URL.Query().Get("time")
	if timeDate == "" {
		http.Error(w, "Missing time URL parameter", http.StatusBadRequest)
		return
	}

	// Split ISO string to get date. ex: 2021-03-01T00:00:00Z -> 2021-03-01
	// Gets put into struct later
	date := strings.Split(timeDate, "T")[0]

	// Mirror request to https://wms.geonorge.no/skwms1/wms.traktorveg_skogsbilveger
	startRequest := time.Now()
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
	log.Printf("Fetching data took: %v", time.Since(startRequest))

	// Group the features by EPSG25833 coordinates, each with a cluster at coordinates: xxx500, yyy500
	// This is a center point of a 1000x1000 meter square, and the center of SeNorge grid cells
	// This is done to reduce the number of requests to the frost API, as the frost API is slow and cannot
	// handle requests with multiple coordinates in the same grid cell

	startProcessing := time.Now()
	shardedMap := clusterWFSResponseToShardedMap(wfsResponse)
	featureMap := shardedMap.getFeaturesFromShardedMap()
	log.Printf("Processing features took: %v", time.Since(startProcessing))

	startFrost := time.Now()
	isFrozenMap, err := mapGridCentersToFrozenStatus(shardedMap.getHashSetFromShardedMap(), date)
	if err != nil {
		http.Error(w, "Failed to get frost data", http.StatusInternalServerError)
		log.Println("Error getting frost data: ", err)
		return
	}
	log.Printf("Getting frost data took: %v", time.Since(startFrost))

	transcribedFeatures := make([]WFSFeature, 0)

	// Iterate over the featuremap and update the features with the frost data
	startUpdate := time.Now()
	for key, features := range featureMap {
		isFrozen, ok := isFrozenMap[key]
		if !ok {
			http.Error(w, "Failed to get frost data", http.StatusInternalServerError)
			log.Println("Error getting frost data: ", err)
			return
		}

		for i := range features {
			features[i].IsFrozen = isFrozen
			transcribedFeatures = append(transcribedFeatures, features[i])
		}
	}
	log.Printf("Updating features took: %v", time.Since(startUpdate))
	log.Printf("Total length of features: %v", len(transcribedFeatures))

	// Test superficial deposits for first feature
	codes, err := GetSuperficialDepositCodesForFeature(transcribedFeatures[0])
	if err != nil {
		http.Error(w, "Failed to get superficial deposit data", http.StatusInternalServerError)
		log.Println("Error getting superficial deposit data: ", err)
		return
	}

	log.Printf("Superficial deposit codes: %v", codes)

	// Classify the features
	for i := range transcribedFeatures {
		ClassifyFeature(&transcribedFeatures[i])
	}

	// Replace the features with the transcribed features
	wfsResponse.Features = transcribedFeatures

	// Encode response
	err = json.NewEncoder(w).Encode(wfsResponse)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Println("Error encoding final response: ", err)
		return
	}

	log.Printf("Total request timeDate: %v", time.Since(startTotal))
}

func ClassifyFeature(feature *WFSFeature) {
	if feature.IsFrozen {
		feature.Properties.Farge = []int{0, 0, 255}
	} else {
		feature.Properties.Farge = []int{255, 0, 0}
	}
}
