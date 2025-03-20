package service

import (
	"github.com/fasthttp/router"
	"github.com/urvin/gokaru/internal/server/helper"
	"github.com/valyala/fasthttp"
	"log/slog"
)

type Handler struct {
	Logger *slog.Logger
}

func (h *Handler) Register(router *router.Router) {
	router.GET("/health", h.health)
	router.GET("/favicon.ico", h.favicon)
	router.NotFound = h.notfound
	router.MethodNotAllowed = h.notallowed
}

func (h *Handler) health(context *fasthttp.RequestCtx) {
	context.SetStatusCode(fasthttp.StatusOK)
	_, err := context.WriteString(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err != nil {
		context.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}

func (h *Handler) favicon(context *fasthttp.RequestCtx) {
	fasthttp.ServeFile(context, "/var/gokaru/assets/favicon.ico")
}

func (h *Handler) notfound(context *fasthttp.RequestCtx) {
	h.Logger.Warn(
		"URL not found",
		"context", "server",
		"handler", "not_found",
		"url", context.URI().String(),
	)
	helper.ServeError(context, fasthttp.StatusNotFound, fasthttp.StatusMessage(fasthttp.StatusNotFound))
}

func (h *Handler) notallowed(context *fasthttp.RequestCtx) {
	h.Logger.Warn(
		"URL not allowed",
		"context", "server",
		"handler", "not_allowed",
		"url", context.URI().String(),
	)
	helper.ServeError(context, fasthttp.StatusMethodNotAllowed, fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed))
}
