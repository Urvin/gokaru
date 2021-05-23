package config

type Config struct {
	Port           int `yaml:"port" envconfig:"GOKARU_PORT" default:"8101"`
	Padding        int `yaml:"padding" default:"10"`
	QualityDefault int `yaml:"quality_default" default:"80"`
	Quality        []struct {
		Format     string `yaml:"format"`
		Quality    int    `yaml:"quality"`
		Conditions []struct {
			From    int `yaml:"from"`
			To      int `yaml:"to"`
			Quality int `yaml:"quality"`
		}
	} `yaml:"quality"`
}
