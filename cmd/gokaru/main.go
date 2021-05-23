package main

import (
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"github.com/urvin/gokaru/internal/security"
	"github.com/urvin/gokaru/internal/storage"
	"github.com/urvin/gokaru/internal/thumbnailer"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	config.Initialize()

	router := mux.NewRouter()

	router.HandleFunc("/health", healthHandler)
	router.HandleFunc("/favicon.ico", faviconHandler).Methods("GET")
	router.HandleFunc("/{sourceType:image|file}/{category}/{filename}", uploadHandler).Methods("PUT")
	router.HandleFunc("/{sourceType:image|file}/{category}/{filename}", removeHandler).Methods("DELETE")
	router.HandleFunc("/{sourceType:image|file}/{category}/{filename}", originHandler).Methods("GET")
	router.HandleFunc("/{sourceType:image|file}/{signature}/{category}/{width:[0-9]+}/{height:[0-9]+}/{cast:[0-9]+}/{filename}", thumbnailHandler).Methods("GET")

	log.Info("Gokaru starting")
	http.ListenAndServe(":"+strconv.Itoa(config.Get().Port), router)
}

func healthHandler(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusOK)
	fmt.Fprintf(response, http.StatusText(http.StatusOK))
}

func faviconHandler(response http.ResponseWriter, request *http.Request) {
	response.WriteHeader(http.StatusOK)
}

func uploadHandler(response http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	sourceType := params["sourceType"]
	fileCategory := params["category"]
	fileName := params["filename"]

	destinationFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, fileName)
	destinationPath := filepath.Dir(destinationFileName)

	err := helper.CreatePathIfNotExists(destinationPath)

	uploadedData, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(response, "Cannot read request body: "+err.Error(), http.StatusBadRequest)
		log.Error("Gokaru.upload: cannot read request body: " + err.Error())
		return
	}

	if sourceType == "image" {
		contentType := http.DetectContentType(uploadedData)

		if contentType != "image/bmp" &&
			contentType != "image/gif" &&
			contentType != "image/webp" &&
			contentType != "image/png" &&
			contentType != "image/jpeg" {
			http.Error(response, "Uploaded file is not an image, "+contentType+" uploaded", http.StatusBadRequest)
			log.Error("Gokaru.upload: uploaded file is not an image: " + err.Error())
			return
		}
	}

	temporaryFile, err := ioutil.TempFile("", "upload")
	if err != nil {
		http.Error(response, "Cannot create a temporary file: "+err.Error(), http.StatusInternalServerError)
		log.Error("Gokaru.upload: Cannot create a temporary file: " + err.Error())
		return
	}

	err = ioutil.WriteFile(temporaryFile.Name(), uploadedData, 0644)
	if err != nil {
		http.Error(response, "Cannot write into temporary file: "+err.Error(), http.StatusInternalServerError)
		log.Error("Gokaru.upload: Cannot write into temporary file: " + err.Error())
		return
	}
	defer os.Remove(temporaryFile.Name())

	err = helper.CopyFile(temporaryFile.Name(), destinationFileName)
	if err != nil {
		http.Error(response, "Cannot write uploaded file into destination: "+err.Error(), http.StatusInternalServerError)
		log.Error("Gokaru.upload: Cannot write uploaded file into destination: " + err.Error())
	} else {
		response.WriteHeader(http.StatusCreated)
	}
}

func removeHandler(response http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	sourceType := params["sourceType"]
	fileCategory := params["category"]
	fileName := params["filename"]

	originFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, fileName)
	defer os.Remove(originFileName)

	if sourceType == "image" {
		thumbnailWildcard := storage.GetStorageImageThumbnailFilename(sourceType, fileCategory, fileName, 0, 0, 0, "*", true)
		defer helper.RemoveByWildcard(thumbnailWildcard)
	}
}

func originHandler(response http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)

	sourceType := params["sourceType"]
	fileCategory := params["category"]
	fileName := params["filename"]

	originFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, fileName)
	http.ServeFile(response, request, originFileName)
}

func thumbnailHandler(response http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)

	sourceType := params["sourceType"]
	fileCategory := params["category"]
	fileName := params["filename"]
	width := helper.Atoi(params["width"])
	height := helper.Atoi(params["height"])
	cast := helper.Atoi(params["cast"])
	signature := params["signature"]

	filenameExtension := strings.ToLower(strings.TrimLeft(filepath.Ext(fileName), "."))
	filenameWithoutExtension := helper.FileNameWithoutExtension(fileName)

	// check signature
	generatedSignature := security.GenerateSignature(sourceType, fileCategory, fileName, width, height, cast)
	if signature != generatedSignature {
		response.WriteHeader(http.StatusForbidden)
		log.Warn("Gokaru.thumbnail: signature mismatch")
		return
	}

	// check for webp acceptance
	if filenameExtension != "webp" {
		httpAccept := request.Header.Get("Accept")
		if strings.Contains(httpAccept, "webp") {
			filenameExtension = "webp"
			response.Header().Set("Vary", "Accept")
		}
	}

	thumbnailFileName := storage.GetStorageImageThumbnailFilename(sourceType, fileCategory, fileName, width, height, cast, filenameExtension, false)

	_, err := os.Stat(thumbnailFileName)
	if os.IsNotExist(err) {
		thumbnailPath := filepath.Dir(thumbnailFileName)
		err := helper.CreatePathIfNotExists(thumbnailPath)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			log.Error("Gokaru.thumbnail: cannot create thumbnail storage path: " + err.Error())
			return
		}

		originFileName := storage.GetStorageOriginFilename(sourceType, fileCategory, filenameWithoutExtension)
		err = thumbnailer.Thumbnail(originFileName, thumbnailFileName, width, height, cast, filenameExtension)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			log.Error("Gokaru.thumbnail: cannot generate thumbnail: " + err.Error())
			return
		}
	} else if err != nil {
		// An error else than file does not exist
		response.WriteHeader(http.StatusInternalServerError)
		log.Error("Gokaru.thumbnail: " + err.Error())
		return
	}

	// thumbnail exists
	http.ServeFile(response, request, thumbnailFileName)
}
