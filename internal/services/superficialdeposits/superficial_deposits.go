package superficialdeposits

import (
	"fmt"
	"runtime"
	"skogkursbachelor/server/internal/models"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// index is a spatial index for the forestry roads
var _index = buildIndex()

var _fjordIndex = buildFjordIndex()

func buildIndex() *models.SpatialIndex {
	shapefiles := []string{
		"data/Losmasse/LosmasseFlate_20240621",
		//"data/Losmasse/LosmasseFlate_20240622",
	}

	return models.ReadShapeFilesAndBuildIndex(shapefiles)
}

func buildFjordIndex() *models.SpatialIndex {
	shapefiles := []string{
		"data/Fjord/fjordkatalogen_omrade",
	}

	return models.ReadShapeFilesAndBuildIndex(shapefiles)
}

func UpdateSuperficialDepositCodes(featureMap *map[string][]models.ForestRoad) error {
	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for key, values := range *featureMap {
		// Get code for key, used for validation for senorge
		sliced := strings.Split(key, ",")
		x, err := strconv.ParseFloat(sliced[0], 64)
		if err != nil {
			log.Error().Msg("Failed to parse float: " + sliced[0])
			continue
		}

		y, err := strconv.ParseFloat(sliced[1], 64)
		if err != nil {
			log.Error().Msg("Failed to parse float: " + sliced[1])
			continue
		}

		isInFjord, err := getIsPointInFjord([]float64{x, y})
		if err != nil {
			log.Error().Msg("Error while checking fjord value: " + err.Error())
		}

		for i := range values {
			wg.Add(1)

			// Reserve a slot
			semaphore <- struct{}{}

			go func(road *models.ForestRoad) {
				defer wg.Done()
				defer func() { <-semaphore }()

				codes, err := getSuperficialDepositCodesForRoad(*road)
				if err != nil {
					log.Warn().Msg("Failed to get superficial deposit codes: " + err.Error())
					return
				}

				if len(codes) == 0 {
					log.Warn().Msg("No superficial deposit codes found for road: " + road.Properties.Vegnummer)
				}

				road.Properties.Løsmassekoder = codes
				road.Properties.Erklyngesenterundervann = isInFjord
			}(&values[i])
		}
	}

	wg.Wait()
	return nil
}

func getSuperficialDepositCodesForRoad(road models.ForestRoad) ([]int, error) {
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

	queryEveryMeter := 10
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

	// Query every x meters
	var codes []int
	for i := 0; i < queryAmount; i += queryEveryIndex {
		// Get the superficial deposit code for the current point
		coordinates := road.Geometry.Coordinates[i]
		codesForPoint, err := getSuperficialDepositCodesForPoint(coordinates)
		if err != nil {
			return nil, err
		}
		for _, code := range codesForPoint {
			// If code is 1 (Løsmasser/berggrunn under vann,uspesifisert), skip
			if code == 1 {
				continue
			}

			if !slices.Contains(codes, code) {
				codes = append(codes, code)
			}
		}
	}

	return codes, nil
}

func getSuperficialDepositCodesForPoint(coordinate []float64) ([]int, error) {
	results, err := models.QuerySpatialIndex(_index, coordinate[0], coordinate[1])
	if err != nil {
		return nil, err
	}

	//if results == nil {
	//	return nil, nil
	//}

	var codes []int

	// Get the superficial deposit code, jordart:xx
	for _, result := range results {
		code, ok := result["jordart"]
		if !ok {
			return nil, fmt.Errorf("no jordart in results for point: " + fmt.Sprintf("%f, %f", coordinate[0], coordinate[1]))
		}
		codeInt, ok := code.(int)
		if !ok {
			return nil, fmt.Errorf("jordart is not an int: " + fmt.Sprintf("%f, %f", coordinate[0], coordinate[1]))
		}
		codes = append(codes, codeInt)
	}

	return codes, nil
}

func getIsPointInFjord(coordinate []float64) (bool, error) {
	results, err := models.QuerySpatialIndex(_fjordIndex, coordinate[0], coordinate[1])
	if err != nil {
		return false, fmt.Errorf("failed to query spatial index: " + err.Error())
	}

	if results == nil {
		// If result is nil -> the point is not in a fjord
		return false, nil
	} else {
		return true, nil
	}
}
