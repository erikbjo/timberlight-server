package forestryroads

import (
	"fmt"
	"log"
	"runtime"
	"slices"
	"strconv"
	"sync"
)

func UpdateSuperficialDepositCodesForFeatureArray(featureArray *[]WFSFeature) error {
	//log.Println("Starting UpdateSuperficialDepositCodesForFeatureArray")
	//defer log.Println("Finished UpdateSuperficialDepositCodesForFeatureArray")

	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for i := 0; i < len(*featureArray); i++ {
		wg.Add(1)

		// Reserve a slot
		semaphore <- struct{}{}

		go func(feature *WFSFeature) {
			defer wg.Done()
			// Release the slot
			defer func() { <-semaphore }()

			codes, err := GetSuperficialDepositCodesForFeature(*feature)
			if err != nil {
				return
			}

			feature.SuperficialDepositCodes = codes
		}(&(*featureArray)[i])
	}

	wg.Wait()
	return nil
}

func GetSuperficialDepositCodesForFeature(feature WFSFeature) ([]int, error) {
	//uniqueID := "" + feature.Properties.Vegnummer + "_" + feature.Properties.Frameter + "_" + feature.Properties.Tilmeter
	//log.Println("Getting superficial deposit codes for feature: ", uniqueID)
	//defer log.Println("Finished getting superficial deposit codes for feature: ", uniqueID)

	var codes []int

	// Get the road length
	roadStart, err := strconv.Atoi(feature.Properties.Frameter)
	if err != nil {
		log.Println("Failed to convert frameter to int: ", feature.Properties.Frameter)
		return nil, err
	}
	roadEnd, err := strconv.Atoi(feature.Properties.Tilmeter)
	if err != nil {
		log.Println("Failed to convert tilmeter to int: ", feature.Properties.Tilmeter)
		return nil, err
	}
	roadLength := roadEnd - roadStart

	if roadLength < 0 {
		log.Println("Road length is negative: ", feature.Properties.Vegnummer)
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
	results, err := QuerySpatialIndex(index, coordinate[0], coordinate[1])
	if err != nil {
		return 0, err
	}

	if results == nil {
		log.Println("No results returned for point: ", coordinate)
		return 0, fmt.Errorf("no results for point")
	}

	// Get the superficial deposit code, jordart:xx
	code, ok := results["jordart"]
	if !ok {
		log.Println("Failed to get jordart from results on point: ", coordinate, " results: ", results)
		return 0, fmt.Errorf("no jordart in results")
	}

	return code.(int), nil
}
