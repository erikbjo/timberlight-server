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
var index *models.SpatialIndex

func init() {
	shapefiles := []string{
		"data/Losmasse/LosmasseFlate_20240621",
		"data/Losmasse/LosmasseFlate_20240622",
	}

	// Build spatial index
	index = models.ReadShapeFilesAndBuildIndex(shapefiles)
	log.Info().Msg("Index built successfully!")
}

func UpdateSuperficialDepositCodesForFeatures(features *[]models.ForestRoad) error {
	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for i := 0; i < len(*features); i++ {
		wg.Add(1)

		// Reserve a slot
		semaphore <- struct{}{}

		go func(feature *models.ForestRoad) {
			defer wg.Done()
			// Release the slot
			defer func() { <-semaphore }()

			codes, err := getSuperficialDepositCodesForFeature(*feature)
			if err != nil {
				log.Error().Msg("Failed to get superficial deposit codes: " + err.Error())
				return
			}

			feature.SuperficialDepositCodes = codes
		}(&(*features)[i])
	}

	wg.Wait()
	return nil
}

func getSuperficialDepositCodesForFeature(feature models.ForestRoad) ([]int, error) {
	var codes []int

	// Get the road length
	roadStart, err := strconv.Atoi(feature.Properties.Frameter)
	if err != nil {
		log.Error().Msg("Failed to convert frameter to int: " + feature.Properties.Frameter)
		return nil, err
	}
	roadEnd, err := strconv.Atoi(feature.Properties.Tilmeter)
	if err != nil {
		log.Error().Msg("Failed to convert tilmeter to int: " + feature.Properties.Tilmeter)
		return nil, err
	}
	roadLength := roadEnd - roadStart

	if roadLength < 0 {
		log.Error().Msg("Road length is negative: " + feature.Properties.Vegnummer)
		return nil, fmt.Errorf("road length is negative " + feature.Properties.Vegnummer)
	}

	queryEveryMeter := 50
	queryAmount := roadLength / queryEveryMeter
	if queryAmount == 0 {
		queryAmount = 1
	}
	queryEveryIndex := len(feature.Geometry.Coordinates) / queryAmount

	if queryEveryIndex == 0 {
		queryEveryIndex = 1
	}

	if queryAmount > len(feature.Geometry.Coordinates) {
		queryAmount = len(feature.Geometry.Coordinates)
	}

	// Query every 50 meters
	for i := 0; i < queryAmount; i += queryEveryIndex {
		// Get the superficial deposit code for the current point
		code, err := getSuperficialDepositCodeForPoint(feature.Geometry.Coordinates[i])
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
	results, err := models.QuerySpatialIndex(index, coordinate[0], coordinate[1])
	if err != nil {
		return 0, err
	}

	if results == nil {
		log.Error().Msg("No results returned for point: " + fmt.Sprintf("%f, %f", coordinate[0], coordinate[1]))
		return 0, fmt.Errorf("no results for point")
	}

	// Get the superficial deposit code, jordart:xx
	code, ok := results["jordart"]
	if !ok {
		log.Error().Msg("Failed to get jordart from results on point: " + fmt.Sprintf("%f, %f", coordinate[0], coordinate[1]) + " results: " + fmt.Sprintf("%v", results))
		return 0, fmt.Errorf("no jordart in results")
	}

	return code.(int), nil
}
