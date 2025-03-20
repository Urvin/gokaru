package server

import (
	"github.com/fasthttp/router"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/server/handler/service"
	"github.com/urvin/gokaru/internal/server/handler/storage"
	"github.com/urvin/gokaru/internal/server/handler/thumbnail"
	"github.com/urvin/gokaru/internal/version"
	"github.com/valyala/fasthttp"
	"log/slog"
	"strconv"
)

type Server interface {
	ListenAndServe() error
}

type server struct {
	version string
	port    int
	logger  *slog.Logger
	router  *router.Router
}

func NewServer(logger *slog.Logger) Server {
	s := &server{}
	s.logger = logger

	return s
}

func (s *server) ListenAndServe() error {
	s.initRouter()

	srv := &fasthttp.Server{
		Name:               "Gokaru v" + version.Version,
		Handler:            s.router.Handler,
		MaxRequestBodySize: config.Get().MaxUploadSize * 1024 * 1024,
	}

	return srv.ListenAndServe(":" + strconv.Itoa(config.Get().Port))
}

func (s *server) initRouter() {
	s.router = router.New()

	serviceHandler := service.Handler{Logger: s.logger}
	serviceHandler.Register(s.router)

	storageHandler := storage.Handler{Logger: s.logger}
	storageHandler.Register(s.router)

	thumbnailHandler := thumbnail.Handler{Logger: s.logger}
	thumbnailHandler.Register(s.router)
}
