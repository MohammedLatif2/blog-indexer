package config

import (
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Hugo struct {
	BaseURL     string `yaml:"base-url"`
	ContentRoot string `yaml:"content-root"`
}

type Elastic struct {
	Base  string
	Index string
	Type  string
}

type Config struct {
	Hugo    Hugo
	Elastic Elastic
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
	config.Hugo.BaseURL = strings.TrimRight(config.Hugo.BaseURL, "/")
	config.Hugo.ContentRoot = strings.TrimRight(config.Hugo.ContentRoot, "/")
	config.Elastic.Base = strings.TrimRight(config.Elastic.Base, "/")
	return config, nil
}
