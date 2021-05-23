package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

var config *Config

func Initialize() (err error) {
	config = &Config{}

	err = readYml(config)
	if err != nil {
		return
	}

	err = readEnv(config)

	return
}

func Get() Config {
	if config == nil {
		panic("Configuration is not initialized")
	}

	return *config
}

func readYml(cfg *Config) (err error) {
	f, er := os.Open("config.yml")
	if er != nil {
		err = er
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	return
}

func readEnv(cfg *Config) (err error) {
	err = envconfig.Process("", cfg)
	return
}
