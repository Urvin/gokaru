package config

type Config struct {
	Port               int    `yaml:"port" envconfig:"GOKARU_PORT" default:"8101"`
	MaxUploadSize      int    `yaml:"max_upload_size" envconfig:"GOKARU_MAX_UPLOAD_SIZE" default:"100"`
	SignatureSalt      string `yaml:"signature_salt" envconfig:"GOKARU_SIGNATURE_SALT" default:"secret"`
	SignatureAlgorithm string `yaml:"signature_algorithm" envconfig:"GOKARU_SIGNATURE_ALGORITHM" default:"murmur"`
	StoragePath        string `yaml:"storage_path" envconfig:"GOKARU_STORAGE_PATH" default:"./storage/"`
	EnforceWebp        bool   `yaml:"enforce_webp" envconfig:"GOKARU_ENFORCE_WEBP" default:"true"`
	Padding            uint   `yaml:"padding" envconfig:"GOKARU_PADDING" default:"10"`
	QualityDefault     uint   `yaml:"quality_default" envconfig:"GOKARU_QUALITY_DEFAULT" default:"80"`
	Quality            []struct {
		Format     string `yaml:"format"`
		Quality    uint   `yaml:"quality"`
		Iterations uint   `yaml:"iterations"  default:"100"`
		Conditions []struct {
			From       uint `yaml:"from"`
			To         uint `yaml:"to"`
			Quality    uint `yaml:"quality"`
			Iterations uint `yaml:"iterations"  default:"100"`
		}
	} `yaml:"quality"`
}
