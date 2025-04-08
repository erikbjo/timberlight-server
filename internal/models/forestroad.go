package models

import (
	"fmt"
	"math"
	"runtime"
	"skogkursbachelor/server/internal/utils"
	"sync"
)

type WFSResponse struct {
	Type          string `json:"type"`
	NumberMatched int    `json:"numberMatched"`
	Name          string `json:"name"`
	Crs           struct {
		Type       string `json:"type"`
		Properties struct {
			Name string `json:"name"`
		} `json:"properties"`
	} `json:"crs"`
	Date     string       `json:"date"`
	Features []ForestRoad `json:"features"`
}

type ForestRoad struct {
	Type                    string  `json:"type"`
	FrostDepth              float64 `json:"frostDepth"`
	WaterSaturation         float64 `json:"waterSaturation"`
	SuperficialDepositCodes []int   `json:"superficialDepositCodes"`
	SoilTemperature54cm     float64 `json:"soilTemperature54cm"`
	Properties              struct {
		Kommunenummer      string `json:"kommunenummer"`
		Vegkategori        string `json:"vegkategori"`
		Vegfase            string `json:"vegfase"`
		Vegnummer          string `json:"vegnummer"`
		Strekningnummer    string `json:"strekningnummer"`
		Delstrekningnummer string `json:"delstrekningnummer"`
		Frameter           string `json:"frameter"`
		Tilmeter           string `json:"tilmeter"`
	} `json:"properties"`
	Geometry struct {
		Type        string      `json:"type"`
		Coordinates [][]float64 `json:"coordinates"`
	} `json:"geometry"`
}

// ClusterWFSResponseToShardedMap processes the features from the WFS response and clusters them into 1000x1000 meter squares.
// Returns a sharded map with the features clustered by coordinates.
func (wfsResponse WFSResponse) ClusterWFSResponseToShardedMap() *ShardedMap {
	featureMap := NewShardedMap(runtime.NumCPU())

	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for _, feature := range wfsResponse.Features {
		// Reserve a slot
		semaphore <- struct{}{}
		wg.Add(1)

		go func(feature ForestRoad) {
			defer wg.Done()
			// Release the slot
			defer func() { <-semaphore }()

			// Use middle index of coordinates as the middle of the road
			middleIndex := len(feature.Geometry.Coordinates) / 2
			coordinates := feature.Geometry.Coordinates[middleIndex]

			// Round the coordinates to the nearest 500 to cluster into 1000x1000 meter squares
			// The center of the square is the center of the SeNorge grid cell
			roundedX := utils.RoundToNearest500(int(math.Round(coordinates[0])))
			roundedY := utils.RoundToNearest500(int(math.Round(coordinates[1])))
			roundedCoordinates := fmt.Sprintf("%d,%d", roundedX, roundedY)

			// Add the feature to the sharded map
			featureMap.Set(roundedCoordinates, feature)
		}(feature)
	}

	wg.Wait()
	return featureMap
}
