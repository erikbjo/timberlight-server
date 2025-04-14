package models

// NVEFMultiPointTimeSeriesRequest represents the request structure for NVEF MultiPoint Time Series.
type NVEFMultiPointTimeSeriesRequest struct {
	Theme            string `json:"Theme"`
	StartDate        string `json:"StartDate"`
	EndDate          string `json:"EndDate"`
	Format           string `json:"Format"`
	MapCoordinateCsv string `json:"MapCoordinateCsv"`
}

// NVEMultiPointTimeSeriesResponse represents the response structure for NVE MultiPoint Time Series.
type NVEMultiPointTimeSeriesResponse struct {
	CellTimeSeries    []cellTimeSeries `json:"CellTimeSeries"`
	Theme             string           `json:"Theme"`
	FullName          interface{}      `json:"FullName"`
	NoDataValue       int              `json:"NoDataValue"`
	StartDate         string           `json:"StartDate"`
	EndDate           string           `json:"EndDate"`
	PrognoseStartDate interface{}      `json:"PrognoseStartDate"`
	Unit              string           `json:"Unit"`
	TimeResolution    int              `json:"TimeResolution"`
}

// cellTimeSeries represents the time series data for a specific cell in the NVE response.
type cellTimeSeries struct {
	X         int       `json:"X"`
	Y         int       `json:"Y"`
	Altitude  int       `json:"Altitude"`
	CellIndex int       `json:"CellIndex"`
	Data      []float64 `json:"Data"`
}
