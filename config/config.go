package config

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Root   string
	ElRoot string
}

func GetConfig() (*Configuration, error) {
	file, _ := os.Open("config/config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		return nil, err
	}
	return &configuration, nil
}
