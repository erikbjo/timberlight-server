package senorge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/models"
	"strings"

	"github.com/rs/zerolog/log"
)

func UpdateFrostDepth(featureMap *map[string][]models.ForestRoad, date string) error {
	coordinatesString, err := createCoordinateString(*featureMap)
	if err != nil {
		return fmt.Errorf("failed to create coordinate string: %v", err)
	}

	if coordinatesString == "" {
		for _, value := range *featureMap {
			for _, feature := range value {
				feature.Properties.Teledybde = 0
			}
		}
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
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Use NVE api to get frost data
	r, err := http.NewRequest(
		http.MethodPost,
		constants.NVEFrostDepthAPI,
		bytes.NewBuffer(bodyJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	r.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch frozen status: %s", resp.Status)
	}

	defer resp.Body.Close()

	// Decode response
	response := models.NVEMultiPointTimeSeriesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(response.CellTimeSeries) == 0 {
		return fmt.Errorf("no data in response")
	}

	for i := range response.CellTimeSeries {
		key := fmt.Sprintf("%d,%d", response.CellTimeSeries[i].X, response.CellTimeSeries[i].Y)
		slice, ok := (*featureMap)[key]
		if !ok {
			log.Warn().Msgf("featureMap does not contain key: %s", key)
		}

		for j := range slice {
			slice[j].Properties.Teledybde = response.CellTimeSeries[i].Data[0]
		}
	}

	return nil
}

func UpdateWaterSaturation(featureMap *map[string][]models.ForestRoad, date string) error {
	coordinatesString, err := createCoordinateString(*featureMap)
	if err != nil {
		return fmt.Errorf("failed to create coordinate string: %v", err)
	}

	if coordinatesString == "" {
		for _, value := range *featureMap {
			for _, feature := range value {
				feature.Properties.Vannmetning = 0
			}
		}
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
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Use NVE api to get frost data
	r, err := http.NewRequest(
		http.MethodPost,
		constants.NVEFrostDepthAPI,
		bytes.NewBuffer(bodyJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	r.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch frozen status: %s", resp.Status)
	}

	defer resp.Body.Close()

	// Decode response
	response := models.NVEMultiPointTimeSeriesResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(response.CellTimeSeries) == 0 {
		return fmt.Errorf("no data in response")
	}

	for i := range response.CellTimeSeries {
		key := fmt.Sprintf("%d,%d", response.CellTimeSeries[i].X, response.CellTimeSeries[i].Y)
		slice, ok := (*featureMap)[key]
		if !ok {
			log.Warn().Msgf("featureMap does not contain key: %s", key)
		}

		for j := range slice {
			slice[j].Properties.Vannmetning = response.CellTimeSeries[i].Data[0]
		}
	}

	return nil
}

func createCoordinateString(featureMap map[string][]models.ForestRoad) (string, error) {
	if len(featureMap) == 0 {
		return "", fmt.Errorf("feature map is empty")
	}

	// Coordinates is in format "X1 Y1, X2 Y2, ..."
	stringBuilder := strings.Builder{}

	for key, array := range featureMap {
		// If feature is on superficial code 1 (Løsmasser/berggrunn under vann,uspesifisert), skip
		if array[0].Properties.Erklyngesenterundervann {
			continue
		}

		stringBuilder.WriteString(strings.Replace(key, ",", " ", -1))
		stringBuilder.WriteString(", ")
	}

	coordinatesString := stringBuilder.String()

	// Remove last comma
	length := len(coordinatesString)
	if length > 0 {
		coordinatesString = coordinatesString[:length-1]
	}

	return coordinatesString, nil
}
