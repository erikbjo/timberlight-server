package utils

import (
	"encoding/json"
	"os"
)

// LoadProxiesFromFile loads proxy configurations from a JSON file. See proxy.json
func LoadProxiesFromFile() (map[string]string, error) {
	data, err := os.ReadFile("proxy.json")
	if err != nil {
		return nil, err
	}

	var config map[string]string
	err = json.Unmarshal(data, &config)
	return config, err
}
