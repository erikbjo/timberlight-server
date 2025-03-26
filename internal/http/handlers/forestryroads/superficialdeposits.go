package forestryroads

import (
	"fmt"
	"strconv"
)

func GetSuperficialDepositCodesForFeature(feature WFSFeature) ([]int, error) {
	var codes []int

	// Get the road length
	roadStart, err := strconv.Atoi(feature.Properties.Frameter)
	if err != nil {
		return nil, err
	}
	roadEnd, err := strconv.Atoi(feature.Properties.Tilmeter)
	if err != nil {
		return nil, err
	}
	roadLength := roadEnd - roadStart

	if roadLength < 0 {
		return nil, fmt.Errorf("road length is negative " + feature.Properties.Vegnummer)
	}

	queryEveryMeter := 50
	queryAmount := roadLength / queryEveryMeter
	if queryAmount == 0 {
		queryAmount = 1
	}
	queryEveryIndex := len(feature.Geometry.Coordinates) / queryAmount

	// Query every 50 meters
	for i := 0; i < queryAmount; i += queryEveryIndex {
		// Get the superficial deposit code for the current point
		code, err := getSuperficialDepositCodeForPoint(feature.Geometry.Coordinates[i])
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}

	return codes, nil
}

func getSuperficialDepositCodeForPoint(coordinate []float64) (int, error) {
	results, err := QuerySpatialIndex(index, coordinate[0], coordinate[1])
	if err != nil {
		return 0, err
	}

	if results == nil {
		return 0, fmt.Errorf("no results for point")
	}

	// Get the superficial deposit code, jordart:xx
	code, ok := results["jordart"]
	if !ok {
		return 0, fmt.Errorf("no jordart in results")
	}

	return code.(int), nil
}
