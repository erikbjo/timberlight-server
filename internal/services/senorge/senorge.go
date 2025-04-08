package senorge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/models"
	"strings"
)

func MapGridCentersToFrozenStatus(featureMap map[string]bool, date string) (map[string]float64, error) {
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
		log.Error().Msg("Failed to create request: " + err.Error())
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Error().Msg("Failed to do request: " + err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	// Decode response
	response := models.NVEMultiPointTimeSeriesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Msg("Failed to decode response: " + err.Error())
		return nil, err
	}

	if len(response.CellTimeSeries) == 0 {
		log.Error().Msg("No data in response")
		return nil, fmt.Errorf("no data in response")
	}

	// Create map of frost depth values
	frostDepthMap := make(map[string]float64)
	for i := range response.CellTimeSeries {
		key := fmt.Sprintf("%d,%d", response.CellTimeSeries[i].X, response.CellTimeSeries[i].Y)
		frostDepthMap[key] = response.CellTimeSeries[i].Data[0]
	}

	return frostDepthMap, nil
}

func MapGridCentersToWaterSaturation(featureMap map[string]bool, date string) (map[string]float64, error) {
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
		log.Error().Msg("Failed to create request: " + err.Error())
		return nil, err
	}

	r.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Error().Msg("Failed to do request: " + err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	// Decode response
	response := models.NVEMultiPointTimeSeriesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Msg("Failed to decode response: " + err.Error())
		return nil, err
	}

	if len(response.CellTimeSeries) == 0 {
		log.Error().Msg("No data in response")
		return nil, fmt.Errorf("no data in response")
	}

	// Create map of frost depth values
	waterSaturationMap := make(map[string]float64)
	for i := range response.CellTimeSeries {
		key := fmt.Sprintf("%d,%d", response.CellTimeSeries[i].X, response.CellTimeSeries[i].Y)
		waterSaturationMap[key] = response.CellTimeSeries[i].Data[0]
	}

	return waterSaturationMap, nil
}
