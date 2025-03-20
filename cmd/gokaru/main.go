package main

import (
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/di"
	"github.com/urvin/gokaru/internal/server"
	"github.com/urvin/gokaru/internal/vips"
)

func main() {
	err := config.Init()
	if err != nil {
		panic(err)
	}

	err = di.Init()
	if err != nil {
		panic(err)
	}

	err = vips.Startup()
	if err != nil {
		di.Logger().Error(
			"vips startup fail",
			"context", "main",
			"error", err.Error(),
		)
		panic(err)
	}
	defer vips.Shutdown()

	http := server.NewServer(di.Logger())
	err = http.ListenAndServe()
	if err != nil {
		di.Logger().Error(
			"http server crashed",
			"context", "main",
			"error", err.Error(),
		)
		panic(err)
	}
}
