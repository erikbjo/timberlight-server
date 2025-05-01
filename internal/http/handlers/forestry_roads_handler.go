package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/models"
	"skogkursbachelor/server/internal/services/senorge"
	"skogkursbachelor/server/internal/services/superficialdeposits"
	"strings"
	"sync"
	_ "sync"
)

// _implementedMethods is a list of the implemented HTTP methods for the status endpoint.
var _implementedMethods = []string{http.MethodGet}

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
				_implementedMethods,
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

	// If there are 0 roads, just return the wfsResponse
	if wfsResponse.NumberMatched == 0 {
		log.Debug().Str("request", r.URL.String()).Msg("No features found in WFS response")
		err = json.NewEncoder(w).Encode(wfsResponse)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Error().Msg("Error encoding final response: " + err.Error())
			return
		}
		return
	}

	// Group the features by EPSG25833 coordinates, each with a cluster at coordinates: xxx500, yyy500
	// This is a center point of a 1000x1000 meter square, and the center of SeNorge grid cells

	shardedMap := wfsResponse.ClusterWFSResponseToShardedMap()
	featureMap := shardedMap.GetFeaturesFromShardedMap()

	// Superficial depositz
	err = superficialdeposits.UpdateSuperficialDepositCodes(&featureMap)
	if err != nil {
		http.Error(w, "Failed to update superficial deposit data", http.StatusInternalServerError)
		log.Error().Msg("Error updating superficial deposit data: " + err.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var err1 error
	go func() {
		defer wg.Done()
		err1 = senorge.UpdateFrostDepth(&featureMap, date)
	}()

	var err2 error
	go func() {
		defer wg.Done()
		err2 = senorge.UpdateWaterSaturation(&featureMap, date)
	}()

	wg.Wait()

	if err1 != nil {
		log.Error().Msg("Error getting frozen status: " + err1.Error())
	}
	if err2 != nil {
		log.Error().Msg("Error getting waterSaturation: " + err2.Error())
	}
	if err1 != nil || err2 != nil {
		http.Error(w, "Error getting external data", http.StatusInternalServerError)
		return
	}

	transcribedFeatures := make([]models.ForestRoad, 0, len(wfsResponse.Features))

	// Iterate over the featuremap and update the features with the frost data
	for _, features := range featureMap {
		for i := range features {
			transcribedFeatures = append(transcribedFeatures, features[i])
		}
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
