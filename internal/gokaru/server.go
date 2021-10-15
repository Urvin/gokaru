package gokaru

import (
	"github.com/fasthttp/router"
	log "github.com/sirupsen/logrus"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
	"github.com/urvin/gokaru/internal/thumbnailer"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type server struct {
	router *router.Router
}

type Server interface {
	Start() error
}

func (s *server) Start() error {
	s.initRouter()
	return fasthttp.ListenAndServe(":"+strconv.Itoa(config.Get().Port), s.router.Handler)
}

func (s *server) initRouter() {
	s.router = router.New()
	s.router.GET("/health", healthHandler)
	s.router.GET("/favicon.ico", faviconHandler)
	s.router.PUT("/{sourceType:^image|file$}/{category}/{filename}", uploadHandler)
	s.router.DELETE("/{sourceType:^image|file$}/{category}/{filename}", removeHandler)
	s.router.GET("/{sourceType:^image|file$}/{category}/{filename}", originHandler)
	s.router.GET("/{sourceType:^image|file$}/{signature}/{category}/{width:[0-9]+}/{height:[0-9]+}/{cast:[0-9]+}/{filename}", thumbnailHandler)
}

func NewServer() Server {
	result := &server{}
	return result
}

func healthHandler(context *fasthttp.RequestCtx) {
	context.SetStatusCode(fasthttp.StatusOK)
	context.WriteString(fasthttp.StatusMessage(fasthttp.StatusOK))
}

func faviconHandler(context *fasthttp.RequestCtx) {
	context.SetStatusCode(fasthttp.StatusOK)
}

func uploadHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)

	destinationFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, fileName)
	destinationPath := filepath.Dir(destinationFileName)

	err := helper.CreatePathIfNotExists(destinationPath)

	uploadedData := context.Request.Body()

	if sourceType == "image" {
		contentType := http.DetectContentType(uploadedData)

		if contentType != "image/bmp" &&
			contentType != "image/gif" &&
			contentType != "image/webp" &&
			contentType != "image/png" &&
			contentType != "image/jpeg" {
			context.Error("Uploaded file is not an image, "+contentType+" uploaded", fasthttp.StatusBadRequest)
			log.Error("Gokaru.upload: uploaded file is not an image: " + err.Error())
			return
		}
	}

	temporaryFile, err := ioutil.TempFile("", "upload")
	if err != nil {
		context.Error("Cannot create a temporary file: "+err.Error(), fasthttp.StatusInternalServerError)
		log.Error("Gokaru.upload: Cannot create a temporary file: " + err.Error())
		return
	}

	err = ioutil.WriteFile(temporaryFile.Name(), uploadedData, 0644)
	if err != nil {
		context.Error("Cannot write into temporary file: "+err.Error(), fasthttp.StatusInternalServerError)
		log.Error("Gokaru.upload: Cannot write into temporary file: " + err.Error())
		return
	}
	defer os.Remove(temporaryFile.Name())

	err = helper.CopyFile(temporaryFile.Name(), destinationFileName)
	if err != nil {
		context.Error("Cannot write uploaded file into destination: "+err.Error(), fasthttp.StatusInternalServerError)
		log.Error("Gokaru.upload: Cannot write uploaded file into destination: " + err.Error())
	} else {
		context.SetStatusCode(fasthttp.StatusCreated)
	}
}

func removeHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)

	originFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, fileName)
	defer os.Remove(originFileName)

	if sourceType == "image" {
		thumbnailWildcard := storage.GetStorageImageThumbnailFilename(sourceType, fileCategory, fileName, 0, 0, 0, "*", true)
		defer helper.RemoveByWildcard(thumbnailWildcard)
	}
}

func originHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)

	originFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, fileName)
	fasthttp.ServeFile(context, originFileName)
}

func thumbnailHandler(context *fasthttp.RequestCtx) {
	sourceType := context.UserValue("sourceType").(string)
	fileCategory := context.UserValue("category").(string)
	fileName := context.UserValue("filename").(string)
	width := helper.Atoi(context.UserValue("width").(string))
	height := helper.Atoi(context.UserValue("height").(string))
	cast := helper.Atoi(context.UserValue("cast").(string))
	signature := context.UserValue("signature").(string)

	filenameExtension := strings.ToLower(strings.TrimLeft(filepath.Ext(fileName), "."))
	filenameWithoutExtension := helper.FileNameWithoutExtension(fileName)

	// check signature
	generatedSignature := security.GenerateSignature(sourceType, fileCategory, fileName, width, height, cast)
	if signature != generatedSignature {
		context.SetStatusCode(fasthttp.StatusForbidden)
		log.Warn("Gokaru.thumbnail: signature mismatch")
		return
	}

	// check for webp acceptance
	if filenameExtension != "webp" {
		httpAccept := string(context.Request.Header.Peek(fasthttp.HeaderAccept))
		if strings.Contains(httpAccept, "webp") {
			filenameExtension = "webp"
			context.Response.Header.Set(fasthttp.HeaderVary, "Accept")
		}
	}

	thumbnailFileName := storage.GetStorageImageThumbnailFilename(sourceType, fileCategory, fileName, width, height, cast, filenameExtension, false)

	_, err := os.Stat(thumbnailFileName)
	if os.IsNotExist(err) {
		thumbnailPath := filepath.Dir(thumbnailFileName)
		err := helper.CreatePathIfNotExists(thumbnailPath)
		if err != nil {
			context.SetStatusCode(fasthttp.StatusInternalServerError)
			log.Error("Gokaru.thumbnail: cannot create thumbnail storage path: " + err.Error())
			return
		}

		originFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, filenameWithoutExtension)
		err = thumbnailer.Thumbnail(originFileName, thumbnailFileName, width, height, cast, filenameExtension)
		if err != nil {
			context.SetStatusCode(fasthttp.StatusInternalServerError)
			log.Error("Gokaru.thumbnail: cannot generate thumbnail: " + err.Error())
			return
		}
	} else if err != nil {
		// An error else than file does not exist
		context.SetStatusCode(fasthttp.StatusInternalServerError)
		log.Error("Gokaru.thumbnail: " + err.Error())
		return
	}

	// thumbnail exists
	fasthttp.ServeFile(context, thumbnailFileName)
}
