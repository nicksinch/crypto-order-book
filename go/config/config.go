package config

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Depth     string   `json:"depth"`
	Pairs     []string `json:"pairs"`
	Snapshots []string `json:"snapshots"`
}

func InitializeConfig() (*Config, error) {
	var appConfig Config
	jsonConfig, err := os.Open("./websockets-config.json")
	if err != nil {
		return nil, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonConfig.Close()

	raw, err := io.ReadAll(jsonConfig)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &appConfig)
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}
