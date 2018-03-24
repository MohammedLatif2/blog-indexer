package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	HugoRoot    string `yaml:"hugo-root"`
	ElasticBase string `yaml:"elastic-base"`
}

func NewConfig(filename string) (*Config, error) {
	// Read file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// Parse Yaml
	config := &Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
