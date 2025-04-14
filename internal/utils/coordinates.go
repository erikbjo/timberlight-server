package utils

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/twpayne/go-proj/v11"
	"math"
	"strconv"
)

var pj25833ToLongLat = get25833ToLongLat()

func get25833ToLongLat() *proj.PJ {
	pj25833ToLongLat, err := proj.NewCRSToCRS("EPSG:"+strconv.Itoa(25833), "EPSG:"+strconv.Itoa(4326), nil)
	if err != nil {
		log.Fatal().Msg("failed to initialize proj CRS")
	}

	return pj25833ToLongLat
}

// TransformCoordinates transforms coordinates from one EPSG to another.
func TransformCoordinates(coordinates []float64, epsgFrom, epsgTo int) (int, int, error) {
	//log.Println("Transforming coordinates: "+strconv.FormatFloat(coordinates[0], 'f', -1, 64)+", "+strconv.FormatFloat(coordinates[1], 'f', -1, 64)+" from EPSG:", epsgFrom, "to EPSG:", epsgTo)

	pj, err := proj.NewCRSToCRS("EPSG:"+strconv.Itoa(epsgFrom), "EPSG:"+strconv.Itoa(epsgTo), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to initialize proj CRS: %v", err.Error())
	}

	oldCoords := proj.NewCoord(coordinates[0], coordinates[1], 0, 0)
	newCoords, err := pj.Forward(oldCoords)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to transform coordinates: %v", err.Error())
	}

	return int(math.Round(newCoords.X())), int(math.Round(newCoords.Y())), nil
}

// Transform25833ToLongLatRoundedToNearest25Deg transforms coordinates from EPSG:25833 to EPSG:4326 (WGS84)
// and rounds them to the nearest 0.25 degrees.
func Transform25833ToLongLatRoundedToNearest25Deg(coordinates []float64) (float64, float64, error) {
	oldCoords := proj.NewCoord(coordinates[0], coordinates[1], 0, 0)
	newCoords, err := pj25833ToLongLat.Forward(oldCoords)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to transform coordinates: %v", err.Error())
	}

	return RoundToNearest25Deg(newCoords.X()), RoundToNearest25Deg(newCoords.Y()), nil
}
