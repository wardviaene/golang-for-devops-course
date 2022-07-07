package server

import "gopkg.in/yaml.v3"

func ReadConfig(bytes []byte) Config {
	var config Config

	// config parsing
	err := yaml.Unmarshal(bytes, &config)
	if err != nil {
		config.LoadError = err
	}
	return config
}
