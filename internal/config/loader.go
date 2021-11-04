package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"os"
)

var config *Config

func Init() (err error) {
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
	f, er := os.Open("/var/gokaru/config.yml")
	if er != nil {
		err = er
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	return
}

func readEnv(cfg *Config) (err error) {
	err = envconfig.Process("", cfg)
	return
}
