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
	Date     string `json:"date"`
	Features []struct {
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
	} `json:"features"`
}
