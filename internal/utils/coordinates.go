package utils

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"skogkursbachelor/server/internal/constants"
	"strconv"
)

type mapTilerTransformationResponse struct {
	TransformerSelectionStrategy string `json:"transformer_selection_strategy"`
	Results                      []struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"results"`
}

func TransformCoordinates(coordinates []float64, epsgFrom, epsgTo int) (int, int, error) {
	r, err := http.NewRequest(
		http.MethodGet,
		constants.MapTilerTransformAPI+strconv.FormatFloat(coordinates[0], 'f', -1, 64)+","+strconv.FormatFloat(coordinates[1], 'f', -1, 64)+".json?s_srs="+strconv.Itoa(epsgFrom)+"&t_srs="+strconv.Itoa(epsgTo)+"&key="+GetMapTilerAPIKey(),
		nil,
	)

	if err != nil {
		log.Println("Error creating request: ", err)
		return -1, -1, err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println("Error doing request: ", err)
		return -1, -1, err
	}

	defer resp.Body.Close()

	// Decode response
	var response mapTilerTransformationResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Println("Error decoding response: ", err)
		return -1, -1, err
	}

	return int(math.Round(response.Results[0].X)), int(math.Round(response.Results[0].Y)), nil
}
