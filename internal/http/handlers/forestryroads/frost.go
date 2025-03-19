package forestryroads

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"skogkursbachelor/server/internal/utils"
)

type MapGridResponse struct {
	ErrorMessage   *string   `json:"ErrorMessage"`
	MapGridValue   []float64 `json:"MapGridValue"`
	Theme          string    `json:"Theme"`
	Unit           string    `json:"Unit"`
	TimeResolution int       `json:"TimeResolution"`
	MASL           int       `json:"MASL"`
	NoDataValue    int       `json:"NoDataValue"`
}

// GetIsGroundFrozen returns true if the ground is frozen, false otherwise
func GetIsGroundFrozen(coordinates []float64, date string) (bool, error) {
	if len(coordinates) != 2 {
		return false, fmt.Errorf("invalid coordinates, expected [longitude, latitude]")
	}

	// Transform coordinates to UTM zone 33N
	utmX, utmY, err := utils.TransformCoordinates(coordinates, 3857, 25833)
	if err != nil {
		return false, fmt.Errorf("failed to transform coordinates: %v", err)
	}

	startDateTime := date + "T00:00:00Z"
	endDateTime := date + "T23:59:59Z"

	// Build the API URL
	apiURL := "https://services.xgeo.no/seNorgeMapAppServices/Services/UIService.svc/GetMapGridInfo"
	queryParams := fmt.Sprintf(`{"x":%d,"y":%d,"id":"gwb_frd","startDateTime":"%s","endDateTime":"%s"}`,
		utmX, utmY, startDateTime, endDateTime)

	fullURL := apiURL + "?request=" + url.QueryEscape(queryParams)

	// Do the request
	resp, err := http.Get(fullURL)
	if err != nil {
		return false, fmt.Errorf("failed to fetch frost data: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse body
	var result MapGridResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Error from API
	if result.ErrorMessage != nil {
		return false, fmt.Errorf("API error: %s", *result.ErrorMessage)
	}

	// Check if we got any data
	if len(result.MapGridValue) == 0 {
		return false, fmt.Errorf("no frost data available for the given coordinates and date")
	}

	log.Println("Received frost data:", result.MapGridValue)

	// Return true if the frost value is less than or equal to 0
	// TODO: Add proper check for frost value with the correct threshold
	threshold := 0.0
	return result.MapGridValue[0] >= threshold, nil
}
