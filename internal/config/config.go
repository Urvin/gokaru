package config

type Config struct {
	Port               int    `yaml:"port" envconfig:"GOKARU_PORT" default:"8101"`
	SignatureSalt      string `yaml:"signature_salt" envconfig:"GOKARU_SIGNATURE_SALT" default:"secret"`
	SignatureAlgorithm string `yaml:"signature_algorithm" envconfig:"GOKARU_SIGNATURE_ALGORITHM" default:"murmur"`
	StoragePath        string `yaml:"storage_path" envconfig:"GOKARU_STORAGE_PATH" default:"./storage/"`
	Padding            int    `yaml:"padding" default:"10"`
	QualityDefault     int    `yaml:"quality_default" default:"80"`
	Quality            []struct {
		Format     string `yaml:"format"`
		Quality    int    `yaml:"quality"`
		QualityMin int    `yaml:"quality_min" default:"0"`
		Iterations int    `yaml:"iterations"  default:"100"`
		Conditions []struct {
			From       int `yaml:"from"`
			To         int `yaml:"to"`
			Quality    int `yaml:"quality"`
			QualityMin int `yaml:"quality_min" default:"0"`
			Iterations int `yaml:"iterations"  default:"100"`
		}
	} `yaml:"quality"`
}
