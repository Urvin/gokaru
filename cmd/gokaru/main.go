package main

import (
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/gokaru"
	"github.com/urvin/gokaru/internal/logging"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
	"github.com/urvin/gokaru/internal/vips"
)

var Version string

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Info("Gokaru started")

	err := config.Init()
	if err != nil {
		logger.Fatal("Could not read config:" + err.Error())
		panic(err)
	}
	logger.Info("Config read")

	err = vips.Startup()
	if err != nil {
		logger.Fatal("Could not start vips:" + err.Error())
		panic(err)
	}
	defer vips.Shutdown()

	server := gokaru.NewServer(Version, getStorage(), getSignatureGenerator(), logger)
	err = server.Start()
	if err != nil {
		logger.Fatal("Gokaru crashed:" + err.Error())
	}
}

func getSignatureGenerator() security.SignatureGenerator {
	var generator security.SignatureGenerator
	if config.Get().SignatureAlgorithm == "md5" {
		generator = security.NewMd5SignatureGenerator()
	} else {
		generator = security.NewMurmurSignatureGenerator()
	}
	generator.SetSalt(config.Get().SignatureSalt)
	return generator
}

func getStorage() storage.Storage {
	return storage.NewFileStorage(config.Get().StoragePath)
}
