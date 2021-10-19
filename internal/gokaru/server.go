package gokaru

import (
	"bytes"
	"github.com/fasthttp/router"
	log "github.com/sirupsen/logrus"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
	"github.com/urvin/gokaru/internal/thumbnailer"
	"github.com/valyala/fasthttp"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const TIME_FORMAT = "Mon, 02 Jan 2006 15:04:05 GMT"

type server struct {
	router             *router.Router
	signatureGenerator security.SignatureGenerator
	storage            storage.Storage
}

func (s *server) Start() error {
	s.initSignatureGenerator()
	s.initStorage()
	s.initRouter()
	return fasthttp.ListenAndServe(":"+strconv.Itoa(config.Get().Port), s.requestMiddleware)
}

func (s *server) initRouter() {
	s.router = router.New()
	s.router.GET("/health", s.healthHandler)
	s.router.GET("/favicon.ico", s.faviconHandler)
	s.router.PUT("/{sourceType:^image|file$}/{category}/{filename}", s.uploadHandler)
	s.router.DELETE("/{sourceType:^image|file$}/{category}/{filename}", s.removeHandler)
	s.router.GET("/{sourceType:^image|file$}/{category}/{filename}", s.originHandler)
	s.router.GET("/{sourceType:^image$}/{signature}/{category}/{width:[0-9]+}/{height:[0-9]+}/{cast:[0-9]+}/{filename}", s.thumbnailHandler)
}

func (s *server) initSignatureGenerator() {
	if config.Get().SignatureAlgorithm == "md5" {
		s.signatureGenerator = security.NewMd5SignatureGenerator()
	} else {
		s.signatureGenerator = security.NewMurmurSignatureGenerator()
	}
	s.signatureGenerator.SetSalt(config.Get().SignatureSalt)
}

func (s *server) initStorage() {
	s.storage = storage.NewFileStorage(config.Get().StoragePath)
}

func (s *server) requestMiddleware(context *fasthttp.RequestCtx) {
	s.router.Handler(context)
	context.Response.Header.Set(fasthttp.HeaderServer, "Gokaru")
}

func NewServer() Server {
	result := &server{}
	return result
}

func (s *server) healthHandler(context *fasthttp.RequestCtx) {
	context.SetStatusCode(fasthttp.StatusOK)
	_, err := context.WriteString(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err != nil {
		log.Error("[server][health] Could not write output: " + err.Error())
	}
}

func (s *server) faviconHandler(context *fasthttp.RequestCtx) {
	fasthttp.ServeFile(context, "./favicon.ico")
	context.SetStatusCode(fasthttp.StatusOK)
}

func (s *server) uploadHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)

	uploadedData := context.Request.Body()

	if sourceType == "image" {
		contentType := http.DetectContentType(uploadedData)

		if contentType != "image/bmp" &&
			contentType != "image/gif" &&
			contentType != "image/webp" &&
			contentType != "image/png" &&
			contentType != "image/jpeg" {
			context.Error("Uploaded file is not an image, "+contentType+" uploaded", fasthttp.StatusBadRequest)
			log.Error("Gokaru.upload: uploaded file is not an image, " + contentType + " given")
			return
		}
	}

	uploadDataReader := bytes.NewReader(uploadedData)
	err := s.storage.Write(sourceType, fileCategory, fileName, uploadDataReader)
	if err != nil {
		context.Error("Could not upload origin", fasthttp.StatusInternalServerError)
		log.Error("[server][upload] Could not upload file: " + err.Error())
		return
	}

	context.SetStatusCode(fasthttp.StatusCreated)
}

func (s *server) removeHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)

	_ = s.storage.Remove(sourceType, fileCategory, fileName)
	context.SetStatusCode(fasthttp.StatusOK)
}

func (s *server) originHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)

	info, err := s.storage.Read(sourceType, fileCategory, fileName)
	if err != nil {
		context.Error("Could not read origin", fasthttp.StatusInternalServerError)
		log.Error("[server][origin] Could not read origin file: " + err.Error())
	}

	s.serveFile(context, info)
}

func (s *server) thumbnailHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)
	width := helper.Atoi(context.UserValue("width").(string))
	height := helper.Atoi(context.UserValue("height").(string))
	cast := helper.Atoi(context.UserValue("cast").(string))
	signature := context.UserValue("signature").(string)

	// check signature
	generatedSignature := s.signatureGenerator.Sign(sourceType, fileCategory, fileName, width, height, cast)
	if signature != generatedSignature {
		context.SetStatusCode(fasthttp.StatusForbidden)
		log.Warn("[server][thumbnail] Signature mismatch")
		return
	}

	thumbnailFileExtension := strings.ToLower(strings.TrimLeft(filepath.Ext(fileName), "."))

	// check for webp acceptance
	if thumbnailFileExtension != "webp" {
		httpAccept := string(context.Request.Header.Peek(fasthttp.HeaderAccept))
		if strings.Contains(httpAccept, "webp") {
			thumbnailFileExtension = "webp"
			context.Response.Header.Set(fasthttp.HeaderVary, "Accept")
		}
	}

	if !s.storage.ThumbnailExists(sourceType, fileCategory, fileName, width, height, cast, thumbnailFileExtension) {
		filenameWithoutExtension := helper.FileNameWithoutExtension(fileName)
		originInfo, err := s.storage.Read(sourceType, fileCategory, filenameWithoutExtension)
		if err != nil {
			context.Error("Could not read origin", fasthttp.StatusInternalServerError)
			log.Error("[server][thumbnail] Could not read origin: " + err.Error())
			return
		}

		thmbnlr := thumbnailer.NewThumbnailer()
		thumbnailData, err := thmbnlr.Thumbnail(originInfo.Reader, width, height, cast, thumbnailFileExtension)
		if err != nil {
			context.Error("Could not read origin", fasthttp.StatusInternalServerError)
			log.Error("[server][thumbnail] Could not create thumbnail: " + err.Error())
			return
		}

		err = s.storage.WriteThumbnail(sourceType, fileCategory, fileName, width, height, cast, thumbnailFileExtension, thumbnailData)
		if err != nil {
			context.Error("Could not save thumbnail", fasthttp.StatusInternalServerError)
			log.Error("[server][thumbnail] Could not save thumbnail: " + err.Error())
			return
		}
	}

	info, err := s.storage.ReadThumbnail(sourceType, fileCategory, fileName, width, height, cast, thumbnailFileExtension)
	if err != nil {
		context.Error("Could not read thumbnail", fasthttp.StatusInternalServerError)
		log.Error("[server][thumbnail] Could not read thumbnail file: " + err.Error())
		return
	}

	s.serveFile(context, info)
}

func (s *server) serveFile(context *fasthttp.RequestCtx, info storage.FileInfo) {

	if !context.IfModifiedSince(info.ModificationTime) {
		context.NotModified()
		return
	}

	context.Response.Header.Set(fasthttp.HeaderContentType, info.ContentType)
	if !info.ModificationTime.IsZero() && !info.ModificationTime.Equal(time.Unix(0, 0)) {
		context.Response.Header.Set(fasthttp.HeaderLastModified, info.ModificationTime.UTC().Format(TIME_FORMAT))
	}

	_, err := io.Copy(context.Response.BodyWriter(), info.Reader)
	if err != nil {
		log.Error("[server][serve] Could not serve file: " + err.Error())
		context.Error("Could not write origin file to output", fasthttp.StatusInternalServerError)
		return
	}

	context.SetStatusCode(fasthttp.StatusOK)
	context.Response.Header.Set(fasthttp.HeaderCacheControl, "max-age=2592000") // 30d
}
