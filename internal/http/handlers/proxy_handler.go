package handlers

import (
	"io"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

type Proxy struct {
	RemoteAddr string
}

func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the remote address
	remoteURL, err := url.Parse(p.RemoteAddr)
	if err != nil {
		log.Error().Msg("Error parsing remote address: " + err.Error())
		http.Error(w, "Invalid remote address", http.StatusInternalServerError)
		return
	}

	// Create the request
	proxyReq, err := http.NewRequest(r.Method, remoteURL.String()+"?"+r.URL.RawQuery, r.Body)
	if err != nil {
		log.Error().Msg("Error creating request: " + err.Error())
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
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
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	// Set the status code and write the response body
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Error().Msg("Error while copying proxy response: " + err.Error())
	}
}
