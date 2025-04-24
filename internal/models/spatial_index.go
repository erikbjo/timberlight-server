package models

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/rtree"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-shapefile"
	"sync"
)

// SpatialIndex represents an R-tree for efficient spatial queries
type SpatialIndex struct {
	tree rtree.RTree
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSpatialIndex creates a new spatial index
func NewSpatialIndex() *SpatialIndex {
	return &SpatialIndex{
		tree: rtree.RTree{},
		data: make(map[string]interface{}),
	}
}

// Insert adds a geometry and its attributes to the spatial index
func (si *SpatialIndex) Insert(minX, minY, maxX, maxY float64, key string, value interface{}) {
	si.tree.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, key)
	si.data[key] = value
}

// Query finds all geometries that intersect with the given bounding box
func (si *SpatialIndex) Query(minX, minY, maxX, maxY float64) []interface{} {
	si.mu.RLock()
	defer si.mu.RUnlock()

	var results []interface{}

	si.tree.Search(
		[2]float64{minX, minY},
		[2]float64{maxX, maxY},
		func(min, max [2]float64, data interface{}) bool {
			// Append the attributes associated with the key
			if key, ok := data.(string); ok {
				results = append(results, si.data[key])
			}
			return true // Continue searching
		},
	)

	return results
}

// ReadShapeFilesAndBuildIndex reads shapefiles and builds a spatial index
func ReadShapeFilesAndBuildIndex(shapefiles []string) *SpatialIndex {
	index := NewSpatialIndex()
	var wg sync.WaitGroup

	for _, file := range shapefiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()

			sf, err := shapefile.Read(f, nil)
			if err != nil {
				log.Error().Msgf("Error reading shape file %s: %s", f, err.Error())
				return
			}
			log.Info().Msgf("Read shapefile: %s, NumRecords: %d, building r-tree...", f, sf.NumRecords())

			for i := 0; i < sf.NumRecords(); i++ {
				attributes, geometry := sf.Record(i)
				switch g := geometry.(type) {
				case *geom.Polygon:
					bbox := g.Bounds()
					key := fmt.Sprintf("%s_%d", f, i)
					index.Insert(bbox.Min(0), bbox.Min(1), bbox.Max(0), bbox.Max(1), key, attributes)
				case *geom.MultiPolygon:
					for j := 0; j < g.NumPolygons(); j++ {
						polygon := g.Polygon(j)
						bbox := polygon.Bounds()
						key := fmt.Sprintf("%s_%d_%d", f, i, j)
						index.Insert(bbox.Min(0), bbox.Min(1), bbox.Max(0), bbox.Max(1), key, attributes)
					}
				case *geom.MultiLineString:
					for j := 0; j < g.NumLineStrings(); j++ {
						line := g.LineString(j)
						bbox := line.Bounds()
						key := fmt.Sprintf("%s_%d_%d", f, i, j)
						index.Insert(bbox.Min(0), bbox.Min(1), bbox.Max(0), bbox.Max(1), key, attributes)
					}
				default:
					log.Warn().Msgf("Unsupported geometry type in %s: %T", f, geometry)
				}
			}
		}(file)
	}

	wg.Wait()
	return index
}

// QuerySpatialIndex checks if a point is inside any geometry in the spatial index
func QuerySpatialIndex(index *SpatialIndex, x, y float64) (map[string]interface{}, error) {
	results := index.Query(x, y, x, y)

	for _, result := range results {
		if attributes, ok := result.(map[string]interface{}); ok {
			return attributes, nil
		} else {
			return nil, fmt.Errorf("unexpected result type %T", result)
		}
	}

	return nil, nil
}
