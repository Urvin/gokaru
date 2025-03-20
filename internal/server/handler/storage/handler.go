package storage

import (
	"github.com/fasthttp/router"
	"github.com/urvin/gokaru/internal/contracts"
	"github.com/urvin/gokaru/internal/di"
	"github.com/urvin/gokaru/internal/server/helper"
	"github.com/urvin/gokaru/internal/storage"
	"github.com/valyala/fasthttp"
	"log/slog"
)

type Handler struct {
	Logger *slog.Logger
}

func (h *Handler) Register(router *router.Router) {
	router.PUT("/{sourceType:^"+contracts.STORAGE_TYPE_FILE+"$}/{category}/{filename}", h.upload)
	router.DELETE("/{sourceType:^"+contracts.STORAGE_TYPE_FILE+"$}/{category}/{filename}", h.remove)
	router.GET("/{sourceType:^"+contracts.STORAGE_TYPE_FILE+"$}/{category}/{filename}", h.origin)

	router.PUT("/{sourceType:^"+contracts.STORAGE_TYPE_IMAGE+"$}/{category}/{filename:^[^\\.]+$}", h.upload)
	router.DELETE("/{sourceType:^"+contracts.STORAGE_TYPE_IMAGE+"$}/{category}/{filename:^[^\\.]+$}", h.remove)
	router.GET("/{sourceType:^"+contracts.STORAGE_TYPE_IMAGE+"$}/{category}/{filename:^[^\\.]+$}", h.origin)
}

func (h *Handler) upload(context *fasthttp.RequestCtx) {
	origin, err := helper.GetOriginInfoFromContext(context)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusBadRequest, "Could not upload origin")
		h.Logger.Error(
			"Invalid input data",
			"context", "server",
			"handler", "upload",
			"error", err.Error(),
		)
		return
	}

	uploadedData := context.Request.Body()

	if origin.Type == contracts.STORAGE_TYPE_IMAGE {
		contentType := helper.MimeByData(uploadedData)

		if contentType != "image/bmp" &&
			contentType != "image/gif" &&
			contentType != "image/webp" &&
			contentType != "image/png" &&
			contentType != "image/jpeg" {
			helper.ServeError(context, fasthttp.StatusBadRequest, "Uploaded file is not an image")
			h.Logger.Error(
				"Uploaded file is not an image",
				"context", "server",
				"handler", "upload",
				"content_type", contentType,
			)
			return
		}
	}

	err = h.storage().Write(origin, uploadedData)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusInternalServerError, "Could not upload origin")
		h.Logger.Error(
			"Could not upload file",
			"context", "server",
			"handler", "upload",
			"error", err.Error(),
		)
		return
	}

	h.Logger.Info(
		"File uploaded",
		"context", "server",
		"handler", "upload",
		"filename", origin.Category+"/"+origin.Name,
	)
	context.SetStatusCode(fasthttp.StatusCreated)
}

func (h *Handler) origin(context *fasthttp.RequestCtx) {
	origin, err := helper.GetOriginInfoFromContext(context)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusNotFound, "Could not read origin")
		h.Logger.Error(
			"Invalid input data",
			"context", "server",
			"handler", "origin",
			"error", err.Error(),
		)
		return
	}

	info, err := h.storage().Read(origin)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusInternalServerError, "Could not return origin")
		h.Logger.Error(
			"Could not return origin",
			"context", "server",
			"handler", "origin",
			"error", err.Error(),
		)
		return
	}

	err = helper.ServeFile(context, info)

	if err != nil {
		helper.ServeError(context, fasthttp.StatusInternalServerError, "Could not return origin")
		h.Logger.Error(
			"Could not return origin",
			"context", "server",
			"handler", "origin",
			"error", err.Error(),
		)
		return
	}
}

func (h *Handler) remove(context *fasthttp.RequestCtx) {
	origin, err := helper.GetOriginInfoFromContext(context)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusBadRequest, "Could not remove file")
		h.Logger.Error(
			"Invalid input data",
			"context", "server",
			"handler", "remove",
			"error", err.Error(),
		)
		return
	}

	err = h.storage().Remove(origin)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusInternalServerError, "Could not remove file")
		h.Logger.Error(
			"Could not remove file",
			"context", "server",
			"handler", "origin",
			"error", err.Error(),
		)
		return
	}

	h.Logger.Info(
		"File deleted",
		"context", "server",
		"handler", "remove",
		"filename", origin.Category+"/"+origin.Name,
	)
	context.SetStatusCode(fasthttp.StatusNoContent)
}

func (h *Handler) storage() storage.Storage {
	return di.Get("storage").(storage.Storage)
}
