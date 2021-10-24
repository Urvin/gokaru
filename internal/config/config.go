package config

type Config struct {
	Port               int    `yaml:"port" envconfig:"GOKARU_PORT" default:"8101"`
	SignatureSalt      string `yaml:"signature_salt" envconfig:"GOKARU_SIGNATURE_SALT" default:"secret"`
	SignatureAlgorithm string `yaml:"signature_algorithm" envconfig:"GOKARU_SIGNATURE_ALGORITHM" default:"murmur"`
	StoragePath        string `yaml:"storage_path" envconfig:"GOKARU_STORAGE_PATH" default:"./storage/"`
	Padding            uint   `yaml:"padding" default:"10"`
	QualityDefault     uint   `yaml:"quality_default" default:"80"`
	Quality            []struct {
		Format     string `yaml:"format"`
		Quality    uint   `yaml:"quality"`
		QualityMin uint   `yaml:"quality_min" default:"0"`
		Iterations uint   `yaml:"iterations"  default:"100"`
		Conditions []struct {
			From       uint `yaml:"from"`
			To         uint `yaml:"to"`
			Quality    uint `yaml:"quality"`
			QualityMin uint `yaml:"quality_min" default:"0"`
			Iterations uint `yaml:"iterations"  default:"100"`
		}
	} `yaml:"quality"`
}
