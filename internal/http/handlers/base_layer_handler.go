package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

// _implementedMethods is a list of the implemented HTTP methods for the status endpoint.
var _implementedMethodsBaseLayer = []string{http.MethodGet}

func BaseLayerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Switch on the HTTP request method
	switch r.Method {
	case http.MethodGet:
		handleBaseLayerGet(w, r)

	default:
		// If the method is not implemented, return an error with the allowed methods
		http.Error(
			w, fmt.Sprintf(
				"REST Method '%s' not supported. Currently only '%v' are supported.", r.Method,
				_implementedMethodsBaseLayer,
			), http.StatusNotImplemented,
		)
		return
	}
}

// handleBaseLayerGet handles GET requests to the base layer endpoint.
func handleBaseLayerGet(w http.ResponseWriter, r *http.Request) {
	topoType := r.PathValue("type")
	abc := r.PathValue("abc")
	z := r.PathValue("z")
	x := r.PathValue("x")
	y := r.PathValue("y")

	var url string
	switch topoType {
	case "topo":
		url = fmt.Sprintf("https://%s.tile.opentopomap.org", abc)
	case "std":
		url = fmt.Sprintf("https://%s.tile.openstreetmap.org", abc)
	default:
		log.Error().Msg("Invalid topo type in base layer request")
		http.Error(w, "Invalid topo type", http.StatusBadRequest)
		return
	}

	for _, v := range []string{z, x, y} {
		if v == "" {
			continue
		} else {
			_, err := strconv.Atoi(v)
			if err != nil {
				log.Error().Msg("Invalid parameter in base layer request: " + v)
				http.Error(w, "Invalid parameter", http.StatusBadRequest)
				return
			}
			url += fmt.Sprintf("/%s", v)
		}
	}
	url += ".png"

	proxyReq, err := http.NewRequest(r.Method, url, nil)
	if err != nil {
		log.Error().Msg("Error creating request: " + err.Error())
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Set(key, value)
		}
	}

	// Make the request
	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		log.Error().Msg("Error making request: " + err.Error())
		http.Error(w, "Failed to fetch data from WMS server", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		if key == "Access-Control-Allow-Origin" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set the status code and write the response body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
