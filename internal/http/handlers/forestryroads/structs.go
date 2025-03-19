package forestryroads

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

type WFSFeature struct {
	Type       string `json:"type"`
	Properties struct {
		Kommunenummer      string `json:"kommunenummer"`
		Vegkategori        string `json:"vegkategori"`
		Vegfase            string `json:"vegfase"`
		Vegnummer          string `json:"vegnummer"`
		Strekningnummer    string `json:"strekningnummer"`
		Delstrekningnummer string `json:"delstrekningnummer"`
		Frameter           string `json:"frameter"`
		Tilmeter           string `json:"tilmeter"`
		Farge              []int  `json:"farge"`
	} `json:"properties"`
	Geometry struct {
		Type        string      `json:"type"`
		Coordinates [][]float64 `json:"coordinates"`
	} `json:"geometry"`
}

type nveFrostDepthRequest struct {
	Theme            string `json:"Theme"`
	StartDate        string `json:"StartDate"`
	EndDate          string `json:"EndDate"`
	Format           string `json:"Format"`
	MapCoordinateCsv string `json:"MapCoordinateCsv"`
}

type nveFrostDepthResponse struct {
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
