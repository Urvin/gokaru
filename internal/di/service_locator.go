package di

import (
	"github.com/sarulabs/di"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"github.com/urvin/gokaru/internal/queue"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
	"github.com/urvin/gokaru/internal/thumbnailer"
	"log/slog"
	"os"
)

var container di.Container

func Init() (err error) {
	builder, err := di.NewBuilder()
	if err != nil {
		return
	}

	err = builder.Add(di.Def{
		Name: "logger",
		Build: func(ctn di.Container) (interface{}, error) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			return logger, nil
		},
		Close: nil,
	})
	if err != nil {
		return
	}

	err = builder.Add(di.Def{
		Name: "storage",
		Build: func(ctn di.Container) (interface{}, error) {
			s := storage.NewFileStorage(config.Get().StoragePath)
			return s, nil
		},
	})
	if err != nil {
		return
	}

	err = builder.Add(di.Def{
		Name: "signature",
		Build: func(ctn di.Container) (interface{}, error) {
			var generator security.SignatureGenerator
			if config.Get().SignatureAlgorithm == "md5" {
				generator = security.NewMd5SignatureGenerator()
			} else {
				generator = security.NewMurmurSignatureGenerator()
			}
			generator.SetSalt(config.Get().SignatureSalt)
			return generator, nil
		},
	})
	if err != nil {
		return
	}

	err = builder.Add(di.Def{
		Name: "queue",
		Build: func(ctn di.Container) (interface{}, error) {
			logger := ctn.Get("logger").(*slog.Logger)
			strg := ctn.Get("storage").(storage.Storage)
			thmbnlr := thumbnailer.NewThumbnailer(logger)

			procs := config.Get().ThumbnailerProcs
			postProcs := config.Get().ThumbnailerPostProcs
			mx := uint(helper.MaxParallelism())

			if procs == 0 {
				procs = mx
			}
			if postProcs == 0 {
				if mx >= 2 {
					postProcs = mx / 2
				} else {
					postProcs = mx
				}
			}

			s := queue.NewQueue(logger, strg, thmbnlr, procs, postProcs)
			return s, nil
		},
	})
	if err != nil {
		return
	}

	container = builder.Build()

	return
}

func Get(name string) interface{} {
	return container.Get(name)
}

func Logger() *slog.Logger {
	logger := container.Get("logger").(*slog.Logger)
	return logger
}
