package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/gokaru"
)

func main() {
	log.Info("Gokaru starting")

	config.Initialize()

	server := gokaru.NewServer()
	err := server.Start()
	if err != nil {
		log.Fatal("Gokaru stopped:" + err.Error())
	}
}
