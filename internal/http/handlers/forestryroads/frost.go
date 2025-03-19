package forestryroads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/utils"
	"strconv"
	"strings"
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

	body := nveFrostDepthRequest{
		Theme:            "gwb_frd",
		StartDate:        date + "T00",
		EndDate:          date + "T00",
		Format:           "json",
		MapCoordinateCsv: fmt.Sprintf("%d %d", utmX, utmY),
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Use NVE api to get frost data
	r, err := http.NewRequest(
		http.MethodPost,
		constants.NVEFrostDepthAPI,
		bytes.NewBuffer(bodyJSON),
	)

	r.Header.Set("Content-Type", "application/json")

	// Print whole request
	log.Println("Request: ", r)

	if err != nil {
		log.Println("Error creating request: ", err)
		return false, err
	}

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println("Error doing request: ", err)
		return false, err
	}

	defer resp.Body.Close()

	// Decode response
	response := nveFrostDepthResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println("Error decoding response: ", err)
		return false, err
	}

	// Return true if the frost value is less than or equal to 0
	// TODO: Add proper check for frost value with the correct threshold
	threshold := 0.0
	return response.CellTimeSeries[0].Data[0] > threshold, nil
}

// Currently only uses one point in the middle of the road
func GetIsGroundFrozenAlongFeature(feature WFSFeature, date string) (bool, error) {
	// Get middle of the road (ish)
	coordinates := new([][]int)
	for i := range feature.Geometry.Coordinates {
		newX, newY, err := utils.TransformCoordinates(feature.Geometry.Coordinates[i], 3857, 25833)
		if err != nil {
			return false, fmt.Errorf("failed to transform coordinates: %v", err)
		}

		*coordinates = append(*coordinates, []int{newX, newY})
	}

	// Create request body
	// Coordinates is in format "X1 Y1, X2 Y2, ..."
	stringBuilder := strings.Builder{}

	length := len(*coordinates)
	for i := range *coordinates {
		stringBuilder.WriteString(strconv.Itoa((*coordinates)[i][0]))
		stringBuilder.WriteString(" ")
		stringBuilder.WriteString(strconv.Itoa((*coordinates)[i][1]))
		if i < length-1 {
			stringBuilder.WriteString(", ")
		}
	}

	// only take start coordinates, as the API fails when multiple coordinates are in the zane gridbox
	stringBuilder.WriteString(fmt.Sprintf("%d %d", (*coordinates)[0][0], (*coordinates)[0][1]))

	coordinatesString := stringBuilder.String()

	body := nveFrostDepthRequest{
		Theme:            "gwb_frd",
		StartDate:        date + "T00",
		EndDate:          date + "T00",
		Format:           "json",
		MapCoordinateCsv: coordinatesString,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Use NVE api to get frost data
	r, err := http.NewRequest(
		http.MethodPost,
		constants.NVEFrostDepthAPI,
		bytes.NewBuffer(bodyJSON),
	)

	r.Header.Set("Content-Type", "application/json")

	// Print whole request
	log.Println("Request: ", r)

	if err != nil {
		log.Println("Error creating request: ", err)
		return false, err
	}

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println("Error doing request: ", err)
		return false, err
	}

	defer resp.Body.Close()

	// Decode response
	response := nveFrostDepthResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println("Error decoding response: ", err)
		return false, err
	}

	// Print response
	log.Println("Received frost data:", response.CellTimeSeries)

	// TODO: Add proper check for frost value with the correct threshold

	return true, nil
}
