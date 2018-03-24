package config

import (
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Elastic struct {
	Base  string
	Index string
	Type  string
}

type Config struct {
	HugoRoot string `yaml:"hugo-root"`
	Elastic  Elastic
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
	config.Elastic.Base = strings.TrimRight(config.Elastic.Base, "/")
	return config, nil
}
