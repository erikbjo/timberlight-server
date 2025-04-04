package structures

import (
	"fmt"
	"math"
	"runtime"
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
	Features []WFSFeature `json:"features"`
}

// ClusterWFSResponseToShardedMap processes the features from the WFS response and clusters them into 1000x1000 meter squares.
// Returns a sharded map with the features clustered by coordinates.
func (wfsResponse WFSResponse) ClusterWFSResponseToShardedMap() *ShardedMap {
	// Sharded map with a reasonable number of shards
	featureMap := NewShardedMap(runtime.NumCPU())

	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for _, feature := range wfsResponse.Features {
		// Reserve a slot
		semaphore <- struct{}{}
		wg.Add(1)

		go func(feature WFSFeature) {
			defer wg.Done()
			// Release the slot
			defer func() { <-semaphore }()

			// Use middle index of coordinates as the middle of the road
			middleIndex := len(feature.Geometry.Coordinates) / 2
			coordinates := feature.Geometry.Coordinates[middleIndex]

			// Round the coordinates to the nearest 500 to cluster into 1000x1000 meter squares
			// The center of the square is the center of the SeNorge grid cell
			roundedX := roundToNearest500(int(math.Round(coordinates[0])))
			roundedY := roundToNearest500(int(math.Round(coordinates[1])))
			roundedCoordinates := fmt.Sprintf("%d,%d", roundedX, roundedY)

			feature.MiddleOfRoad25833 = []int{roundedX, roundedY}

			// Add the feature to the sharded map
			featureMap.Set(roundedCoordinates, feature)
		}(feature)
	}

	wg.Wait()
	return featureMap
}

// roundToNearest500 rounds a number to the nearest 500.
// This is used to cluster the features into 1000x1000 meter squares, ending in 500.
func roundToNearest500(n int) int {
	base := (n / 1000) * 1000
	return base + 500
}

type WFSFeature struct {
	Type                    string  `json:"type"`
	IsFrozen                bool    `json:"isFrozen"`
	MiddleOfRoad25833       []int   `json:"middleOfRoad25833"`
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
