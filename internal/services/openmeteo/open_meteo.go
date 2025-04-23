package openmeteo

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/models"
	"skogkursbachelor/server/internal/utils"
	"strconv"
	"strings"
)

func UpdateDeepSoilTemp(featureMap *map[string][]models.ForestRoad, date string) error {
	// Cluster the coordinates into 25 degree cells
	clusteredMap := clusterCoordinates(*featureMap)
	soilTempMapLatLong := make(map[string]float64)

	lats := make([]float64, len(clusteredMap))
	lons := make([]float64, len(clusteredMap))

	i := 0
	for k := range clusteredMap {
		lat := strings.Split(k, ",")[0]
		lon := strings.Split(k, ",")[1]

		latFloat, err := strconv.ParseFloat(lat, 64)
		if err != nil {
			log.Error().Msg("Failed to parse latitude: " + lat)
			continue
		}

		lonFloat, err := strconv.ParseFloat(lon, 64)
		if err != nil {
			log.Error().Msg("Failed to parse longitude: " + lon)
			continue
		}

		lats[i] = latFloat
		lons[i] = lonFloat
		i++
	}

	url := *new(string)
	// If date is earlier than today, use the Open Meteo API
	if isBefore, err := utils.IsEarlierThanToday(date); err != nil {
		return fmt.Errorf("failed to check if date is earlier than today: %w", err)
	} else if isBefore {
		url = constants.OpenMeteoHistoricalDeepSoilTempURL
	} else {
		url = constants.OpenMeteoDeepSoilTempURL
	}

	url = strings.Replace(url, "{latitude}", strings.Trim(strings.Replace(fmt.Sprint(lats), " ", ",", -1), "[]"), 1)
	url = strings.Replace(url, "{longitude}", strings.Trim(strings.Replace(fmt.Sprint(lons), " ", ",", -1), "[]"), 1)
	url = strings.Replace(url, "{start_date}", date, 1)
	url = strings.Replace(url, "{end_date}", date, 1)

	log.Info().Str("url", url).Msg("Constructed Open Meteo URL")

	// Get the soil moisture data
	r, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	r.Header.Set("Content-Type", "application/json")

	// Do the request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	defer resp.Body.Close()

	// Decode response
	var response []models.OpenMeteoDeepSoilTempResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("failed to decode response from Open Meteo: %w", err)
	}

	for _, meteoResp := range response {
		soilTempMapLatLong[strconv.FormatFloat(meteoResp.Latitude, 'f', -1, 64)+","+strconv.FormatFloat(meteoResp.Longitude, 'f', -1, 64)] = meteoResp.Hourly.SoilTemperature54Cm[len(meteoResp.Hourly.SoilTemperature54Cm)/2]
	}

	for key, coordinates := range clusteredMap {
		soilTemp := soilTempMapLatLong[key]
		for _, coordinate := range coordinates {
			features := (*featureMap)[coordinate]
			for i := range features {
				features[i].Properties.Jordtemperatur54cm = soilTemp
			}
		}
	}

	return nil
}

func clusterCoordinates(featureMap map[string][]models.ForestRoad) map[string][]string {
	// Create a map with 25 degree cells as keys and the coordinates in the cell as values
	clusteredMap := make(map[string][]string)

	for key := range featureMap {
		// Get the coordinates
		coordinates := strings.Split(key, ",")
		x, _ := strconv.ParseFloat(coordinates[0], 64)
		y, _ := strconv.ParseFloat(coordinates[1], 64)

		newX, newY, err := utils.Transform25833ToLongLatRoundedToNearest25Deg([]float64{x, y})
		if err != nil {
			log.Error().Msg("Error transforming latitude to longitude")
			continue
		}

		// Add the coordinates to the cell
		cellKey := strconv.FormatFloat(newX, 'f', -1, 64) + "," + strconv.FormatFloat(newY, 'f', -1, 64)
		clusteredMap[cellKey] = append(clusteredMap[cellKey], key)
	}

	return clusteredMap
}
