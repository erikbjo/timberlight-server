package models

// OpenMeteoDeepSoilTempResponse represents the response structure from the Open Meteo API for deep soil temperature.
type OpenMeteoDeepSoilTempResponse struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	HourlyUnits struct {
		Time                string `json:"time"`
		SoilTemperature54Cm string `json:"soil_temperature_54cm"`
	} `json:"hourly_units"`
	Hourly struct {
		Time                []string  `json:"time"`
		SoilTemperature54Cm []float64 `json:"soil_temperature_54cm"`
	} `json:"hourly"`
	LocationId int `json:"location_id,omitempty"`
}
