package storage

import (
	"github.com/urvin/gokaru/internal/helper"
	"path/filepath"
	"strconv"
)

const IMAGE_ORIGIN_PATH = "origin"
const IMAGE_THUMBNAIL_PATH = "thumbnail"

func GetStoragePath() string {
	result, _ := filepath.Abs("./storage/")
	return result
}

func GetStorageOriginFilename(sourceType, fileCategory, fileName string) string {
	hashedFileName := helper.MD5(fileName)
	hashedFilePath := helper.Substring(hashedFileName, 0, 2) + "/" + helper.Substring(hashedFileName, 2, 4)

	result := GetStoragePath() + "/" + sourceType
	if sourceType == "image" {
		result = result + "/" + IMAGE_ORIGIN_PATH
	}
	result = result + "/" + fileCategory + "/" + hashedFilePath + "/" + hashedFileName

	return result
}

func GetStorageImageThumbnailFilename(sourceType, fileCategory, fileName string, width, height, cast int, extension string, del bool) string {
	hashedFileName := helper.MD5(fileName)
	hashedFilePath := helper.Substring(hashedFileName, 0, 2) + "/" + helper.Substring(hashedFileName, 2, 4)

	castPath := "*"
	extensionPart := "*"
	if !del {
		castPath = strconv.Itoa(width) + "x" + strconv.Itoa(height) + "x" + strconv.Itoa(cast)
		extensionPart = extension
	}

	result := GetStoragePath() + "/" + sourceType + "/" + IMAGE_THUMBNAIL_PATH + "/"
	result = result + "/" + fileCategory + "/" + castPath + "/" + hashedFilePath + "/" + hashedFileName + "." + extensionPart

	return result
}
