package main

import (
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/gokaru"
	"github.com/urvin/gokaru/internal/logging"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Info("Gokaru started")

	err := config.Init()
	if err != nil {
		logger.Fatal("Could not read config:" + err.Error())
	}
	logger.Info("Config read")

	imagick.Initialize()
	defer imagick.Terminate()

	server := gokaru.NewServer(getStorage(), getSignatureGenerator(), logger)
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
