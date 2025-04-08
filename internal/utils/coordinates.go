package utils

import (
	"github.com/twpayne/go-proj/v11"
	"log"
	"math"
	"strconv"
)

var pj25833ToLongLat *proj.PJ

func init() {
	var err error
	pj25833ToLongLat, err = proj.NewCRSToCRS("EPSG:"+strconv.Itoa(25833), "EPSG:"+strconv.Itoa(4326), nil)
	if err != nil {
		log.Fatal("Failed to projection CRSToCRS:", err)
	}
}

func TransformCoordinates(coordinates []float64, epsgFrom, epsgTo int) (int, int, error) {
	//log.Println("Transforming coordinates: "+strconv.FormatFloat(coordinates[0], 'f', -1, 64)+", "+strconv.FormatFloat(coordinates[1], 'f', -1, 64)+" from EPSG:", epsgFrom, "to EPSG:", epsgTo)

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

func Transform25833ToLongLatRoundedToNearest25Deg(coordinates []float64) (float64, float64, error) {
	oldCoords := proj.NewCoord(coordinates[0], coordinates[1], 0, 0)
	newCoords, err := pj25833ToLongLat.Forward(oldCoords)
	if err != nil {
		log.Println("Failed to transform coordinates:", err)
		return 0, 0, err
	}

	return RoundToNearest25Deg(newCoords.X()), RoundToNearest25Deg(newCoords.Y()), nil
}
