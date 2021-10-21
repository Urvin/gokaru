package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/gokaru"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
)

func main() {
	log.Info("Gokaru starting")

	err := config.Initialize()
	if err != nil {
		log.Fatal("Could not read config:" + err.Error())
	}

	server := gokaru.NewServer(getStorage(), getSignatureGenerator())
	err = server.Start()
	if err != nil {
		log.Fatal("Gokaru stopped:" + err.Error())
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
