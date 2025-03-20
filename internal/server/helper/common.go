package helper

import (
	"encoding/json"
	"errors"
	"github.com/urvin/gokaru/internal/contracts"
	"github.com/urvin/gokaru/internal/helper"
	"github.com/valyala/fasthttp"
	"time"
)

const TIME_FORMAT = "Mon, 02 Jan 2006 15:04:05 GMT"

func WriteJsonContent(context *fasthttp.RequestCtx, model interface{}) (err error) {
	content, err := json.Marshal(model)
	if err != nil {
		return
	}

	context.SetStatusCode(fasthttp.StatusCreated)
	context.SetContentType("application/json; charset=utf-8")
	context.SetBody(content)

	return
}

func GetOriginInfoFromContext(context *fasthttp.RequestCtx) (origin *contracts.OriginDto, err error) {
	origin = &contracts.OriginDto{
		Type:     context.UserValue("sourceType").(string),
		Category: context.UserValue("category").(string),
		Name:     context.UserValue("filename").(string),
	}
	if len(origin.Type) == 0 {
		err = errors.New("type is empty")
	}
	if len(origin.Category) == 0 {
		err = errors.New("category is empty")
	}
	if len(origin.Name) == 0 {
		err = errors.New("name is empty")
	}
	return
}

func GetMiniatureInfoFromContext(context *fasthttp.RequestCtx) (miniature *contracts.MiniatureDto, err error) {

	miniature = &contracts.MiniatureDto{
		Type:     context.UserValue("sourceType").(string),
		Category: context.UserValue("category").(string),
		Width:    helper.Atoi(context.UserValue("width").(string)),
		Height:   helper.Atoi(context.UserValue("height").(string)),
		Cast:     helper.Atoi(context.UserValue("cast").(string)),
	}
	if len(miniature.Type) == 0 {
		err = errors.New("type is empty")
	}
	if len(miniature.Category) == 0 {
		err = errors.New("category is empty")
	}
	if miniature.Width < 0 {
		err = errors.New("width should not be less tan 0")
	}
	if miniature.Height < 0 {
		err = errors.New("height should not be less tan 0")
	}
	if miniature.Cast < 0 {
		err = errors.New("cast should not be less tan 0")
	}

	filename := context.UserValue("filename").(string)
	miniature.Name = helper.FileNameWithoutExtension(filename)
	miniature.Extension = helper.FileNameExtension(filename)

	if len(miniature.Name) == 0 {
		err = errors.New("name is empty")
	}
	if len(miniature.Extension) == 0 {
		err = errors.New("extension is empty")
	}
	return
}

func ServeFile(context *fasthttp.RequestCtx, info contracts.FileDto) (err error) {

	if !context.IfModifiedSince(info.ModificationTime) {
		context.NotModified()
		return
	}

	context.Response.Header.Set(fasthttp.HeaderContentType, info.ContentType)
	if !info.ModificationTime.IsZero() && !info.ModificationTime.Equal(time.Unix(0, 0)) {
		context.Response.Header.Set(fasthttp.HeaderLastModified, info.ModificationTime.UTC().Format(TIME_FORMAT))
	}

	_, err = context.Write(info.Contents)
	if err != nil {
		return
	}

	context.SetStatusCode(fasthttp.StatusOK)
	context.Response.Header.Set(fasthttp.HeaderCacheControl, "max-age=2592000") // 30d
	return
}

func ServeBytes(context *fasthttp.RequestCtx, data []byte, extension string) (err error) {
	contentType := MimeByExtension(extension)
	if contentType == "" {
		contentType = MimeByData(data)
	}

	context.Response.Header.Set(fasthttp.HeaderContentType, contentType)
	context.Response.Header.Set(fasthttp.HeaderLastModified, time.Now().UTC().Format(TIME_FORMAT))

	_, err = context.Write(data)

	return
}
