package structures

type OpenMeteoResponse struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	HourlyUnits struct {
		Time                 string `json:"time"`
		SoilMoisture10To40Cm string `json:"soil_moisture_10_to_40cm"`
		SoilMoisture0To10Cm  string `json:"soil_moisture_0_to_10cm"`
	} `json:"hourly_units"`
	Hourly struct {
		Time                 []string  `json:"time"`
		SoilMoisture10To40Cm []float64 `json:"soil_moisture_10_to_40cm"`
		SoilMoisture0To10Cm  []float64 `json:"soil_moisture_0_to_10cm"`
	} `json:"hourly"`
}

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
