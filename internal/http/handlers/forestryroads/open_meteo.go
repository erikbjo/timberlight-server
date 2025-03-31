package forestryroads

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"skogkursbachelor/server/internal/http/handlers/forestryroads/structures"
	"skogkursbachelor/server/internal/utils"
	"strconv"
	"strings"
)

func mapGridCentersToSoilMoisture(gridCentersMap map[string]bool, date string) (map[string]float64, error) {
	// Cluster the coordinates into 25 degree cells
	clusteredMap := clusterCoordinates(gridCentersMap)
	soilMoistureMapLatLong := make(map[string]float64)

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

	// One request
	url := constants.OpenMeteoEnsembleAPI
	url = strings.Replace(url, "{latitude}", strings.Trim(strings.Replace(fmt.Sprint(lats), " ", ",", -1), "[]"), 1)
	url = strings.Replace(url, "{longitude}", strings.Trim(strings.Replace(fmt.Sprint(lons), " ", ",", -1), "[]"), 1)
	url = strings.Replace(url, "{start_date}", date, 1)
	url = strings.Replace(url, "{end_date}", date, 1)

	log.Info().Str("url", url).Msg("Open Meteo Ensemble")

	// Get the soil moisture data
	r, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
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
	var response []structures.OpenMeteoResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error().Msg("Failed to decode response: " + err.Error())
		return nil, err
	}

	for _, meteoResp := range response {
		soilMoistureMapLatLong[strconv.FormatFloat(meteoResp.Latitude, 'f', -1, 64)+","+strconv.FormatFloat(meteoResp.Longitude, 'f', -1, 64)] = meteoResp.Hourly.SoilMoisture10To40Cm[len(meteoResp.Hourly.SoilMoisture10To40Cm)/2]
	}

	soilmoistureMap25833 := make(map[string]float64)

	for key, coordinates := range clusteredMap {
		soilMoisture := soilMoistureMapLatLong[key]
		for _, coordinate := range coordinates {
			soilmoistureMap25833[coordinate] = soilMoisture
		}
	}

	return soilmoistureMap25833, nil
}

func clusterCoordinates(featureMap map[string]bool) map[string][]string {
	// Create a map with 25 degree cells as keys and the coordinates in the cell as values
	clusteredMap := make(map[string][]string)

	for key := range featureMap {
		// Get the coordinates
		coordinates := strings.Split(key, ",")
		x, _ := strconv.ParseFloat(coordinates[0], 64)
		y, _ := strconv.ParseFloat(coordinates[1], 64)

		newX, newY, err := utils.Transform25833ToLongLatRoundedToNearest50Deg([]float64{x, y})
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
