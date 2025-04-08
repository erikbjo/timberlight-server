package models

type NVEFMultiPointTimeSeriesRequest struct {
	Theme            string `json:"Theme"`
	StartDate        string `json:"StartDate"`
	EndDate          string `json:"EndDate"`
	Format           string `json:"Format"`
	MapCoordinateCsv string `json:"MapCoordinateCsv"`
}

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

type cellTimeSeries struct {
	X         int       `json:"X"`
	Y         int       `json:"Y"`
	Altitude  int       `json:"Altitude"`
	CellIndex int       `json:"CellIndex"`
	Data      []float64 `json:"Data"`
}
