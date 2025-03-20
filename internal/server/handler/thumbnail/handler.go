package thumbnail

import (
	"github.com/fasthttp/router"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/contracts"
	"github.com/urvin/gokaru/internal/di"
	"github.com/urvin/gokaru/internal/queue"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/server/helper"
	"github.com/valyala/fasthttp"
	"log/slog"
	"strings"
)

type Handler struct {
	Logger *slog.Logger
}

func (h *Handler) Register(router *router.Router) {
	router.GET("/{sourceType:^"+contracts.STORAGE_TYPE_IMAGE+"$}/{signature}/{category}/{width:[0-9]+}/{height:[0-9]+}/{cast:[0-9]+}/{filename}", h.thumbnail)
}

func (h *Handler) thumbnail(context *fasthttp.RequestCtx) {
	miniature, err := helper.GetMiniatureInfoFromContext(context)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusBadRequest, "Could not thumbnail origin")
		h.Logger.Error(
			"Invalid input data",
			"context", "server",
			"handler", "thumbnail",
			"error", err.Error(),
		)
		return
	}

	sg := di.Get("signature").(security.SignatureGenerator)

	signature := context.UserValue("signature").(string)

	// check signature
	generatedSignature := sg.Sign(miniature)
	if signature != generatedSignature {
		helper.ServeError(context, fasthttp.StatusForbidden, "Signature mismatch")
		h.Logger.Warn(
			"Signature mismatch",
			"context", "server",
			"handler", "thumbnail",
			"signature", signature,
			"calculated_signature", generatedSignature,
		)
		return
	}

	if config.Get().EnforceWebp && miniature.Extension != "webp" {
		httpAccept := string(context.Request.Header.Peek(fasthttp.HeaderAccept))
		if strings.Contains(httpAccept, "webp") {
			miniature.Extension = "webp"
			context.Response.Header.Set(fasthttp.HeaderVary, "Accept")
		}
	}

	q := di.Get("queue").(*queue.Queue)
	thumbnail, err := q.GetThumbnail(miniature)
	if err != nil {
		helper.ServeError(context, fasthttp.StatusInternalServerError, "Could not process thumbnail")
		h.Logger.Error(
			"Thumbnail processing error",
			"context", "server",
			"handler", "thumbnail",
			"error", err.Error(),
		)
		return
	}

	if thumbnail.Size == 0 {
		err = helper.ServeBytes(context, thumbnail.Contents, miniature.Extension)
	} else {
		err = helper.ServeFile(context, thumbnail)
	}

	if err != nil {
		h.Logger.Error(
			"Thumbnail showing error",
			"context", "server",
			"handler", "thumbnail",
			"error", err.Error(),
		)
	}

	return
}
