package senorge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/models"
	"strings"
)

func MapGridCentersToFrozenStatus(featureMap map[string]bool, date string) (map[string]float64, error) {
	coordinatesString, err := createCoordinateString(featureMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinate string: %v", err)
	}

	body := models.NVEFMultiPointTimeSeriesRequest{
		Theme:            constants.SeNorgeFrostDepthTheme,
		StartDate:        date + "T12",
		EndDate:          date + "T12",
		Format:           "json",
		MapCoordinateCsv: coordinatesString,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Use NVE api to get frost data
	r, err := http.NewRequest(
		http.MethodPost,
		constants.NVEFrostDepthAPI,
		bytes.NewBuffer(bodyJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	r.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %v", err)
	}

	defer resp.Body.Close()

	// Decode response
	response := models.NVEMultiPointTimeSeriesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(response.CellTimeSeries) == 0 {
		return nil, fmt.Errorf("no data in response")
	}

	// Create map of frost depth values
	frostDepthMap := make(map[string]float64, len(response.CellTimeSeries))
	for i := range response.CellTimeSeries {
		key := fmt.Sprintf("%d,%d", response.CellTimeSeries[i].X, response.CellTimeSeries[i].Y)
		frostDepthMap[key] = response.CellTimeSeries[i].Data[0]
	}

	return frostDepthMap, nil
}

func MapGridCentersToWaterSaturation(featureMap map[string]bool, date string) (map[string]float64, error) {

	coordinatesString, err := createCoordinateString(featureMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinate string: %v", err)
	}

	body := models.NVEFMultiPointTimeSeriesRequest{
		Theme:            constants.SeNorgeWaterSaturationTheme,
		StartDate:        date + "T12",
		EndDate:          date + "T12",
		Format:           "json",
		MapCoordinateCsv: coordinatesString,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Use NVE api to get frost data
	r, err := http.NewRequest(
		http.MethodPost,
		constants.NVEFrostDepthAPI,
		bytes.NewBuffer(bodyJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	r.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %v", err)
	}

	defer resp.Body.Close()

	// Decode response
	response := models.NVEMultiPointTimeSeriesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if len(response.CellTimeSeries) == 0 {
		return nil, fmt.Errorf("no data in response")
	}

	// Create map of frost depth values
	waterSaturationMap := make(map[string]float64, len(response.CellTimeSeries))
	for i := range response.CellTimeSeries {
		key := fmt.Sprintf("%d,%d", response.CellTimeSeries[i].X, response.CellTimeSeries[i].Y)
		waterSaturationMap[key] = response.CellTimeSeries[i].Data[0]
	}

	return waterSaturationMap, nil
}

func createCoordinateString(featureMap map[string]bool) (string, error) {
	if len(featureMap) == 0 {
		return "", fmt.Errorf("feature map is empty")
	}

	// Coordinates is in format "X1 Y1, X2 Y2, ..."
	stringBuilder := strings.Builder{}

	for key := range featureMap {
		stringBuilder.WriteString(strings.Replace(key, ",", " ", -1))
		stringBuilder.WriteString(",")
	}

	coordinatesString := stringBuilder.String()

	// Remove last comma
	length := len(coordinatesString)
	if length > 0 {
		coordinatesString = coordinatesString[:length-1]
	}

	return coordinatesString, nil
}
