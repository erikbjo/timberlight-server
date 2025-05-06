package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// _implementedMethods is a list of the implemented HTTP methods for the status endpoint.
var _implementedMethodsLegend = []string{http.MethodGet}

// ForestryRoadsHandler handles requests to the forestry road endpoint.
// Currently only GET requests are supported.
func ForestryLegendHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleForestryLegendGet(w, r)

	default:
		http.Error(
			w, fmt.Sprintf(
				"REST Method '%s' not supported. Currently only '%v' are supported.", r.Method,
				_implementedMethodsLegend,
			), http.StatusNotImplemented,
		)
		return
	}
}

// handleForestryLegendGet handles GET requests to the forestry legend endpoint.
func handleForestryLegendGet(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("assets", "forestry_road_legend.png")

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Could not open image file", http.StatusInternalServerError)
		log.Error().Msg("Failed to open forestry road legend" + err.Error())
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "image/png")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to send image", http.StatusInternalServerError)
		log.Error().Msg("Failed to send forestry road legend" + err.Error())
	}
}
