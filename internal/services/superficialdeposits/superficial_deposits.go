package superficialdeposits

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"runtime"
	"skogkursbachelor/server/internal/models"
	"slices"
	"strconv"
	"sync"
)

// index is a spatial index for the forestry roads
var _index = buildIndex()

func buildIndex() *models.SpatialIndex {
	shapefiles := []string{
		"data/Losmasse/LosmasseFlate_20240621",
		"data/Losmasse/LosmasseFlate_20240622",
	}

	return models.ReadShapeFilesAndBuildIndex(shapefiles)
}

func UpdateSuperficialDepositCodes(roads *[]models.ForestRoad) error {
	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for i := 0; i < len(*roads); i++ {
		wg.Add(1)

		// Reserve a slot
		semaphore <- struct{}{}

		go func(road *models.ForestRoad) {
			defer wg.Done()
			// Release the slot
			defer func() { <-semaphore }()

			codes, err := getSuperficialDepositCodesForRoad(*road)
			if err != nil {
				log.Warn().Msg("Failed to get superficial deposit codes: " + err.Error())
				return
			}

			road.Properties.Løsmassekoder = codes
		}(&(*roads)[i])
	}

	wg.Wait()
	return nil
}

func getSuperficialDepositCodesForRoad(road models.ForestRoad) ([]int, error) {
	var codes []int

	// Get the road length
	roadStart, err := strconv.Atoi(road.Properties.Frameter)
	if err != nil {
		return nil, fmt.Errorf("failed to convert frameter to int: " + err.Error())
	}
	roadEnd, err := strconv.Atoi(road.Properties.Tilmeter)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tilmeter to int: " + err.Error())
	}
	roadLength := roadEnd - roadStart

	if roadLength < 0 {
		return nil, fmt.Errorf("road length is negative " + road.Properties.Vegnummer)
	}

	queryEveryMeter := 50
	queryAmount := roadLength / queryEveryMeter
	if queryAmount == 0 {
		queryAmount = 1
	}
	queryEveryIndex := len(road.Geometry.Coordinates) / queryAmount

	if queryEveryIndex == 0 {
		queryEveryIndex = 1
	}

	if queryAmount > len(road.Geometry.Coordinates) {
		queryAmount = len(road.Geometry.Coordinates)
	}

	// Query every 50 meters
	// TODO: goroutine?
	for i := 0; i < queryAmount; i += queryEveryIndex {
		// Get the superficial deposit code for the current point
		code, err := getSuperficialDepositCodeForPoint(road.Geometry.Coordinates[i])
		if err != nil {
			return nil, err
		}
		if !slices.Contains(codes, code) {
			codes = append(codes, code)
		}
	}

	return codes, nil
}

func getSuperficialDepositCodeForPoint(coordinate []float64) (int, error) {
	results, err := models.QuerySpatialIndex(_index, coordinate[0], coordinate[1])
	if err != nil {
		return 0, err
	}

	if results == nil {
		return 0, fmt.Errorf("no results returned for point: " + fmt.Sprintf("%f, %f", coordinate[0], coordinate[1]))
	}

	// Get the superficial deposit code, jordart:xx
	code, ok := results["jordart"]
	if !ok {
		return 0, fmt.Errorf("no jordart in results for point: " + fmt.Sprintf("%f, %f", coordinate[0], coordinate[1]))
	}

	return code.(int), nil
}
