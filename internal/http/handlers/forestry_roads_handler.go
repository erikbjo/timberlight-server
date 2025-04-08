package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/models"
	"skogkursbachelor/server/internal/services/openmeteo"
	"skogkursbachelor/server/internal/services/senorge"
	"skogkursbachelor/server/internal/services/superficialdeposits"
	"strings"
)

// implementedMethods is a list of the implemented HTTP methods for the status endpoint.
var implementedMethods = []string{http.MethodGet}

// ForestryRoadsHandler handles requests to the forestry road endpoint.
// Currently only GET requests are supported.
func ForestryRoadsHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
		log.Warn().Str("request", r.URL.String()).Msg("Missing time URL parameter")
		http.Error(w, "Missing time URL parameter", http.StatusBadRequest)
		return
	}

	// Split ISO string to get date. ex: 2021-03-01T00:00:00Z -> 2021-03-01
	// Gets put into struct later
	date := strings.Split(timeDate, "T")[0]
	if date == "" {
		log.Warn().Str("request", r.URL.String()).Msg("Failed to split time string")
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
	var wfsResponse models.WFSResponse
	err = json.NewDecoder(proxyResp.Body).Decode(&wfsResponse)
	if err != nil {
		http.Error(w, "Failed to decode external response", http.StatusInternalServerError)
		log.Error().Msg("Error decoding response from GeoNorge WMS server: " + err.Error())
		return
	}

	// Group the features by EPSG25833 coordinates, each with a cluster at coordinates: xxx500, yyy500
	// This is a center point of a 1000x1000 meter square, and the center of SeNorge grid cells

	shardedMap := wfsResponse.ClusterWFSResponseToShardedMap()
	featureMap := shardedMap.GetFeaturesFromShardedMap()

	frostDepthMap, err := senorge.MapGridCentersToFrozenStatus(shardedMap.GetHashSetFromShardedMap(), date)
	if err != nil {
		http.Error(w, "Failed to get frost data", http.StatusInternalServerError)
		log.Error().Msg("Error getting frost data: " + err.Error())
		return
	}

	waterSaturationMap, err := senorge.MapGridCentersToWaterSaturation(shardedMap.GetHashSetFromShardedMap(), date)
	if err != nil {
		http.Error(w, "Failed to get water saturation data", http.StatusInternalServerError)
		log.Error().Msg("Error getting water saturation data: " + err.Error())
		return
	}

	soilTempMap, err := openmeteo.MapGridCentersToDeepSoilTemp(shardedMap.GetHashSetFromShardedMap(), date)
	if err != nil {
		http.Error(w, "Failed to get soil temperature data", http.StatusInternalServerError)
		log.Error().Msg("Error getting soil temperature data: " + err.Error())
		return
	}

	transcribedFeatures := make([]models.ForestRoad, 0)

	// Iterate over the featuremap and update the features with the frost data
	for key, features := range featureMap {
		frostDepth, ok := frostDepthMap[key]
		if !ok {
			http.Error(w, "Failed to get frost data", http.StatusInternalServerError)
			log.Error().Msg("Error getting frost data from frostDepthMap, key: " + key)
			// Not returning here, as we want to update the features with the frost data we have
		}

		waterSaturation, ok := waterSaturationMap[key]
		if !ok {
			http.Error(w, "Failed to get water saturation data", http.StatusInternalServerError)
			log.Error().Msg("Error getting water saturation data from waterSaturationMap, key: " + key)
		}

		soilTemp, ok := soilTempMap[key]
		if !ok {
			http.Error(w, "Failed to get soil temperature data", http.StatusInternalServerError)
			log.Error().Msg("Error getting soil temperature data from soilTempMap, key: " + key)
		}

		for i := range features {
			features[i].FrostDepth = frostDepth
			features[i].WaterSaturation = waterSaturation
			features[i].SoilTemperature54cm = soilTemp
			transcribedFeatures = append(transcribedFeatures, features[i])
		}
	}

	err = superficialdeposits.UpdateSuperficialDepositCodesForFeatures(&transcribedFeatures)
	if err != nil {
		http.Error(w, "Failed to update superficial deposit data", http.StatusInternalServerError)
		log.Error().Msg("Error updating superficial deposit data: " + err.Error())
		return
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
