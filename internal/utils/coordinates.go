package utils

import (
	"github.com/twpayne/go-proj/v11"
	"log"
	"math"
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
	log.Println("Transforming coordinates: "+strconv.FormatFloat(coordinates[0], 'f', -1, 64)+", "+strconv.FormatFloat(coordinates[1], 'f', -1, 64)+" from EPSG:", epsgFrom, "to EPSG:", epsgTo)

	pj, err := proj.NewCRSToCRS("EPSG:"+strconv.Itoa(epsgFrom), "EPSG:"+strconv.Itoa(epsgTo), nil)
	if err != nil {
		log.Println("Failed to projection CRSToCRS:", err)
		return 0, 0, err
	}

	oldCoords := proj.NewCoord(coordinates[0], coordinates[1], 0, 0)
	newCoords, err := pj.Forward(oldCoords)
	if err != nil {
		log.Println("Failed to transform coordinates:", err)
		return 0, 0, err
	}

	return int(math.Round(newCoords.X())), int(math.Round(newCoords.Y())), nil
}
