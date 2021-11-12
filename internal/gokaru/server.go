package gokaru

import (
	"errors"
	"github.com/fasthttp/router"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/contracts"
	"github.com/urvin/gokaru/internal/fileinfo"
	"github.com/urvin/gokaru/internal/helper"
	"github.com/urvin/gokaru/internal/logging"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
	"github.com/urvin/gokaru/internal/thumbnailer"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const TIME_FORMAT = "Mon, 02 Jan 2006 15:04:05 GMT"

type server struct {
	version            string
	router             *router.Router
	signatureGenerator security.SignatureGenerator
	storage            storage.Storage
	logger             logging.Logger
	thumbnailer        thumbnailer.Thumbnailer
}

func (s *server) Start() error {
	s.initRouter()

	server := &fasthttp.Server{
		Name:               s.version,
		Handler:            s.router.Handler,
		MaxRequestBodySize: config.Get().MaxUploadSize * 1024 * 1024,
	}

	return server.ListenAndServe(":" + strconv.Itoa(config.Get().Port))
}

func (s *server) initRouter() {
	s.router = router.New()
	s.router.NotFound = s.notFoundHandler

	s.router.GET("/health", s.healthHandler)
	s.router.GET("/favicon.ico", s.faviconHandler)

	// files can contain dots in filename
	s.router.PUT("/{sourceType:^"+contracts.TYPE_FILE+"$}/{category}/{filename}", s.uploadHandler)
	s.router.DELETE("/{sourceType:^"+contracts.TYPE_FILE+"$}/{category}/{filename}", s.removeHandler)
	s.router.GET("/{sourceType:^"+contracts.TYPE_FILE+"$}/{category}/{filename}", s.originHandler)

	// images cannot contain dots in filename
	s.router.PUT("/{sourceType:^"+contracts.TYPE_IMAGE+"$}/{category}/{filename:^[^\\.]+$}", s.uploadHandler)
	s.router.DELETE("/{sourceType:^"+contracts.TYPE_IMAGE+"$}/{category}/{filename:^[^\\.]+$}", s.removeHandler)
	s.router.GET("/{sourceType:^"+contracts.TYPE_IMAGE+"$}/{category}/{filename:^[^\\.]+$}", s.originHandler)
	s.router.GET("/{sourceType:^"+contracts.TYPE_IMAGE+"$}/{signature}/{category}/{width:[0-9]+}/{height:[0-9]+}/{cast:[0-9]+}/{filename}", s.thumbnailHandler)
}

func (s *server) initStorage() {
	s.storage = storage.NewFileStorage(config.Get().StoragePath)
}

func (s *server) healthHandler(context *fasthttp.RequestCtx) {
	context.SetStatusCode(fasthttp.StatusOK)
	_, err := context.WriteString(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err != nil {
		context.SetStatusCode(fasthttp.StatusInternalServerError)
		s.logger.Error("[server][health] Could not write output: " + err.Error())
	}
}

func (s *server) faviconHandler(context *fasthttp.RequestCtx) {
	fasthttp.ServeFile(context, "/var/gokaru/favicon.ico")
}

func (s *server) notFoundHandler(context *fasthttp.RequestCtx) {
	s.logger.Warn("[server] Path not found: " + string(context.URI().Path()))
	s.serveError(context, fasthttp.StatusNotFound, fasthttp.StatusMessage(fasthttp.StatusNotFound))
}

func (s *server) uploadHandler(context *fasthttp.RequestCtx) {
	origin, err := s.getOriginInfoFromContext(context)
	if err != nil {
		s.serveError(context, fasthttp.StatusBadRequest, "Could not upload origin")
		s.logger.Error("[server][upload] Invalid input data: " + err.Error())
		return
	}

	uploadedData := context.Request.Body()

	if origin.Type == contracts.TYPE_IMAGE {
		contentType := fileinfo.MimeByData(uploadedData)

		if contentType != "image/bmp" &&
			contentType != "image/gif" &&
			contentType != "image/webp" &&
			contentType != "image/png" &&
			contentType != "image/jpeg" {
			s.serveError(context, fasthttp.StatusBadRequest, "Uploaded file is not an image")
			s.logger.Error("Gokaru.upload: uploaded file is not an image, " + contentType + " given")
			return
		}
	}

	err = s.storage.Write(origin, uploadedData)
	if err != nil {
		s.serveError(context, fasthttp.StatusInternalServerError, "Could not upload origin")
		s.logger.Error("[server][upload] Could not upload file: " + err.Error())
		return
	}

	s.logger.Info("[server][upload] File uploaded: " + origin.Category + "/" + origin.Name)
	context.SetStatusCode(fasthttp.StatusCreated)
}

func (s *server) removeHandler(context *fasthttp.RequestCtx) {
	origin, err := s.getOriginInfoFromContext(context)
	if err != nil {
		s.serveError(context, fasthttp.StatusBadRequest, "Could not remove files")
		s.logger.Error("[server][remove] Invalid input data: " + err.Error())
		return
	}

	_ = s.storage.Remove(origin)
	s.logger.Info("[server][remove] File removed: " + origin.Category + "/" + origin.Name)
	context.SetStatusCode(fasthttp.StatusOK)
}

func (s *server) originHandler(context *fasthttp.RequestCtx) {
	origin, err := s.getOriginInfoFromContext(context)
	if err != nil {
		s.serveError(context, fasthttp.StatusBadRequest, "Could not read origin")
		s.logger.Error("[server][origin] Invalid input data: " + err.Error())
		return
	}

	info, err := s.storage.Read(origin)
	if err != nil {
		s.serveError(context, fasthttp.StatusInternalServerError, "Could not read origin")
		s.logger.Error("[server][origin] Could not read origin file: " + err.Error())
	}

	s.serveFile(context, info)
}

func (s *server) thumbnailHandler(context *fasthttp.RequestCtx) {

	miniature, err := s.getMiniatureInfoFromContext(context)
	if err != nil {
		s.serveError(context, fasthttp.StatusBadRequest, "Could not thumbnail origin")
		s.logger.Error("[server][thumbnail] Invalid input data: " + err.Error())
		return
	}

	signature := context.UserValue("signature").(string)

	// check signature
	generatedSignature := s.signatureGenerator.Sign(miniature)
	if signature != generatedSignature {
		s.serveError(context, fasthttp.StatusForbidden, "Signature mismatch")
		s.logger.Warn("[server][thumbnail] Signature mismatch")
		return
	}

	thumbnailFileExtension := strings.ToLower(strings.TrimLeft(filepath.Ext(miniature.Name), "."))
	filenameWithoutExtension := helper.FileNameWithoutExtension(miniature.Name)
	miniature.Name = filenameWithoutExtension

	if config.Get().EnforceWebp && thumbnailFileExtension != "webp" {
		httpAccept := string(context.Request.Header.Peek(fasthttp.HeaderAccept))
		if strings.Contains(httpAccept, "webp") {
			thumbnailFileExtension = "webp"
			context.Response.Header.Set(fasthttp.HeaderVary, "Accept")
		}
	}

	if !s.storage.ThumbnailExists(miniature, thumbnailFileExtension) {
		origin := contracts.Origin{
			Type:     miniature.Type,
			Category: miniature.Category,
			Name:     miniature.Name,
		}
		originInfo, err := s.storage.Read(&origin)
		if err != nil {
			s.serveError(context, fasthttp.StatusInternalServerError, "Could not read origin")
			s.logger.Error("[server][thumbnail] Could not read origin: " + err.Error())
			return
		}

		to := thumbnailer.ThumbnailOptions{}
		to.SetWidth(uint(miniature.Width))
		to.SetHeight(uint(miniature.Height))
		to.SetImageTypeWithExtension(thumbnailFileExtension)
		to.SetOptionsWithCast(uint(miniature.Cast))

		thumbnailData, later, err := s.thumbnailer.Thumbnail(originInfo.Contents, to)
		if err != nil {
			s.serveError(context, fasthttp.StatusInternalServerError, "Could not create thumbnail")
			s.logger.Error("[server][thumbnail] Could not create thumbnail: " + err.Error())
			return
		}

		err = s.storage.WriteThumbnail(miniature, thumbnailFileExtension, thumbnailData)
		if err != nil {
			s.serveError(context, fasthttp.StatusInternalServerError, "Could not save thumbnail")
			s.logger.Error("[server][thumbnail] Could not save thumbnail: " + err.Error())
			return
		}

		s.serveBytes(context, thumbnailData, thumbnailFileExtension)

		if later != nil {
			go func() {
				info, er := s.storage.ReadThumbnail(miniature, thumbnailFileExtension)
				if er != nil {
					s.logger.Error("[server][later] Could not read thumbnail: " + er.Error())
					return
				}

				data, er := later(info.Contents)
				if er != nil {
					s.logger.Error("[server][later] Could not later thumbnail: " + er.Error())
					return
				}

				er = s.storage.WriteThumbnail(miniature, thumbnailFileExtension, data)
				if er != nil {
					s.logger.Error("[server][later] Could not save thumbnail: " + er.Error())
					return
				}
			}()
		}

	} else {
		info, err := s.storage.ReadThumbnail(miniature, thumbnailFileExtension)
		if err != nil {
			s.serveError(context, fasthttp.StatusInternalServerError, "Could not read thumbnail")
			s.logger.Error("[server][thumbnail] Could not read thumbnail file: " + err.Error())
			return
		}

		s.serveFile(context, info)
	}
}

func (s *server) serveFile(context *fasthttp.RequestCtx, info contracts.File) {

	if !context.IfModifiedSince(info.ModificationTime) {
		context.NotModified()
		return
	}

	context.Response.Header.Set(fasthttp.HeaderContentType, info.ContentType)
	if !info.ModificationTime.IsZero() && !info.ModificationTime.Equal(time.Unix(0, 0)) {
		context.Response.Header.Set(fasthttp.HeaderLastModified, info.ModificationTime.UTC().Format(TIME_FORMAT))
	}

	_, err := context.Write(info.Contents)
	if err != nil {
		s.logger.Error("[server][serve] Could not serve file: " + err.Error())
		context.Error("Could not write origin file to output", fasthttp.StatusInternalServerError)
		return
	}

	context.SetStatusCode(fasthttp.StatusOK)
	context.Response.Header.Set(fasthttp.HeaderCacheControl, "max-age=2592000") // 30d
}

func (s *server) serveBytes(context *fasthttp.RequestCtx, data []byte, extension string) {
	contentType := fileinfo.MimeByExtension(extension)
	if contentType == "" {
		contentType = fileinfo.MimeByData(data)
	}

	context.Response.Header.Set(fasthttp.HeaderContentType, contentType)
	context.Response.Header.Set(fasthttp.HeaderLastModified, time.Now().UTC().Format(TIME_FORMAT))
	_, err := context.Write(data)
	if err != nil {
		s.logger.Error("[server][serve] Could not serve file: " + err.Error())
		context.Error("Could not write origin file to output", fasthttp.StatusInternalServerError)
		return
	}
}

func (s *server) getOriginInfoFromContext(context *fasthttp.RequestCtx) (origin *contracts.Origin, err error) {
	origin = &contracts.Origin{
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

func (s *server) getMiniatureInfoFromContext(context *fasthttp.RequestCtx) (miniature *contracts.Miniature, err error) {
	miniature = &contracts.Miniature{
		Type:     context.UserValue("sourceType").(string),
		Category: context.UserValue("category").(string),
		Name:     context.UserValue("filename").(string),
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
	if len(miniature.Name) == 0 {
		err = errors.New("name is empty")
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
	return
}

func (s *server) serveError(context *fasthttp.RequestCtx, statusCode int, title string) {
	context.Response.Reset()
	context.SetStatusCode(statusCode)

	httpAccept := string(context.Request.Header.Peek(fasthttp.HeaderAccept))
	if strings.Contains(httpAccept, "text/html") {
		tb, err := ioutil.ReadFile("/var/gokaru/error.html")
		if err != nil {
			s.logger.Error("[server][serve] Could not read file: " + err.Error())
			return
		}

		ts := string(tb)
		ts = strings.Replace(ts, "#CODE#", strconv.Itoa(statusCode), -1)
		ts = strings.Replace(ts, "#TITLE#", title, -1)
		context.SetContentType("text/html; charset=utf-8")
		context.SetBodyString(ts)
	} else {
		context.SetContentType("text/plain; charset=utf-8")
		context.SetBodyString(title)
	}
}

func NewServer(v string, st storage.Storage, g security.SignatureGenerator, l logging.Logger) Server {
	s := &server{}
	s.version = "Gokaru v" + v
	s.signatureGenerator = g
	s.storage = st
	s.logger = l
	s.thumbnailer = thumbnailer.NewThumbnailer(l)
	return s
}
