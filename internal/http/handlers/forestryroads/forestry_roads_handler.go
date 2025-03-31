package forestryroads

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/http/handlers/forestryroads/structures"
	"strings"
)

// implementedMethods is a list of the implemented HTTP methods for the status endpoint.
var implementedMethods = []string{http.MethodGet}

// index is a spatial index for the forestry roads
var index *structures.SpatialIndex

func init() {
	shapefiles := []string{
		"data/Losmasse/LosmasseFlate_20240621",
		"data/Losmasse/LosmasseFlate_20240622",
	}

	// Build spatial index
	index = structures.ReadShapeFilesAndBuildIndex(shapefiles)
	log.Info().Msg("Index built successfully!")

	log.Info().Msg("Forestry roads handler initialized")
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
	// Get timeDate parameter from url
	timeDate := r.URL.Query().Get("time")
	if timeDate == "" {
		http.Error(w, "Missing time URL parameter", http.StatusBadRequest)
		return
	}

	// Split ISO string to get date. ex: 2021-03-01T00:00:00Z -> 2021-03-01
	// Gets put into struct later
	date := strings.Split(timeDate, "T")[0]
	if date == "" {
		http.Error(w, "Failed to split time string", http.StatusBadRequest)
		return
	}

	// Mirror request to https://wms.geonorge.no/skwms1/wms.traktorveg_skogsbilveger
	proxyReq, err := http.NewRequest(
		r.Method,
		constants.ForestryRoadsWFS+"?"+r.URL.RawQuery,
		r.Body,
	)
	if err != nil {
		http.Error(w, "Failed to create internal request", http.StatusInternalServerError)
		log.Error().Msg("Error creating request to GeoNorge for forestry roads: " + err.Error())
		return
	}

	// Do request
	proxyResp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to fetch data from external WMS server", http.StatusBadGateway)
		log.Error().Msg("Error fetching data from GeoNorge WMS server: " + err.Error())
		return
	}

	// Decode into struct
	var wfsResponse structures.WFSResponse
	err = json.NewDecoder(proxyResp.Body).Decode(&wfsResponse)
	if err != nil {
		http.Error(w, "Failed to decode external response", http.StatusInternalServerError)
		log.Error().Msg("Error decoding response from GeoNorge WMS server: " + err.Error())
		return
	}

	// Group the features by EPSG25833 coordinates, each with a cluster at coordinates: xxx500, yyy500
	// This is a center point of a 1000x1000 meter square, and the center of SeNorge grid cells
	// This is done to reduce the number of requests to the frost API, as the frost API is slow and cannot
	// handle requests with multiple coordinates in the same grid cell

	shardedMap := wfsResponse.ClusterWFSResponseToShardedMap()
	featureMap := shardedMap.GetFeaturesFromShardedMap()

	isFrozenMap, err := mapGridCentersToFrozenStatus(shardedMap.GetHashSetFromShardedMap(), date)
	if err != nil {
		http.Error(w, "Failed to get frost data", http.StatusInternalServerError)
		log.Error().Msg("Error getting frost data: " + err.Error())
		return
	}

	soilMoistureMap, err := mapGridCentersToSoilMoisture(shardedMap.GetHashSetFromShardedMap(), date)
	if err != nil {
		http.Error(w, "Failed to get soil moisture data", http.StatusInternalServerError)
		log.Error().Msg("Error getting soil moisture data: " + err.Error())
		return
	}

	transcribedFeatures := make([]structures.WFSFeature, 0)

	// Iterate over the featuremap and update the features with the frost data
	for key, features := range featureMap {
		isFrozen, ok := isFrozenMap[key]
		if !ok {
			http.Error(w, "Failed to get frost data", http.StatusInternalServerError)
			log.Error().Msg("Error getting frost data from isFrozenMap, key: " + key)
			// Not returning here, as we want to update the features with the frost data we have
		}

		soilMoisture, ok := soilMoistureMap[key]
		if !ok {
			http.Error(w, "Failed to get soil moisture data", http.StatusInternalServerError)
			log.Error().Msg("Error getting soil moisture data from soilMoistureMap, key: " + key)
			// Not returning here, as we want to update the features with the soil moisture data we have
		}

		for i := range features {
			features[i].IsFrozen = isFrozen
			features[i].SoilMoisture10cm40cm = soilMoisture
			transcribedFeatures = append(transcribedFeatures, features[i])
		}
	}

	err = updateSuperficialDepositCodesForFeatureArray(&transcribedFeatures)
	if err != nil {
		http.Error(w, "Failed to update superficial deposit data", http.StatusInternalServerError)
		log.Error().Msg("Error updating superficial deposit data: " + err.Error())
		return
	}

	// Classify the features
	for i := range transcribedFeatures {
		classifyFeature(&transcribedFeatures[i])
	}

	// Replace the features with the transcribed features
	wfsResponse.Features = transcribedFeatures

	// Encode response
	err = json.NewEncoder(w).Encode(wfsResponse)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Error().Msg("Error encoding final response: " + err.Error())
		return
	}
}

func classifyFeature(feature *structures.WFSFeature) {
	if feature.IsFrozen {
		feature.Properties.Farge = []int{0, 0, 255}
	} else {
		feature.Properties.Farge = []int{255, 0, 0}
	}
}
