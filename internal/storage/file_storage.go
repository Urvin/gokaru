package storage

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/n-marshall/go-cp"
	"github.com/urvin/gokaru/internal/contracts"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const IMAGE_ORIGIN_PATH = "origin"
const IMAGE_THUMBNAIL_PATH = "thumbnail"

type fileStorage struct {
	storagePath string
}

func (fs *fileStorage) createPathIfNotExists(path string) (err error) {
	if _, er := os.Stat(path); os.IsNotExist(er) {
		err = os.MkdirAll(path, os.ModePerm)
	}
	return
}

func (fs *fileStorage) removeByWildcard(wildcard string) (err error) {
	files, er := filepath.Glob(wildcard)
	if er != nil {
		err = er
		return
	}
	for _, f := range files {
		if er := os.Remove(f); er != nil {
			err = er
			return
		}
	}
	return
}

func (fs *fileStorage) hashFileName(fileName string) string {
	hash := md5.Sum([]byte(fileName))
	return hex.EncodeToString(hash[:])
}

func (fs *fileStorage) getOriginFilename(origin *contracts.Origin) string {
	hashedFileName := fs.hashFileName(origin.Name)
	hashedFilePath := hashedFileName[0:2] + "/" + hashedFileName[2:4]

	result := fs.storagePath + "/" + origin.Type
	if origin.Type == contracts.TYPE_IMAGE {
		result = result + "/" + IMAGE_ORIGIN_PATH
	}
	result = result + "/" + origin.Category + "/" + hashedFilePath + "/" + hashedFileName

	return result
}

func (fs *fileStorage) getImageThumbnailFilename(miniature *contracts.Miniature, extension string, del bool) string {
	hashedFileName := fs.hashFileName(miniature.Name)
	hashedFilePath := hashedFileName[0:2] + "/" + hashedFileName[2:4]

	castPath := "*"
	extensionPart := "*"
	if !del {
		castPath = strconv.Itoa(miniature.Width) + "x" + strconv.Itoa(miniature.Height) + "x" + strconv.Itoa(miniature.Cast)
		extensionPart = extension
	}

	result := fs.storagePath + "/" + miniature.Type + "/" + IMAGE_THUMBNAIL_PATH
	result = result + "/" + miniature.Category + "/" + castPath + "/" + hashedFilePath + "/" + hashedFileName + "." + extensionPart

	return result
}

func (fs *fileStorage) SetStoragePath(path string) {
	fs.storagePath, _ = filepath.Abs(path)
}

func (fs *fileStorage) Write(origin *contracts.Origin, data []byte) (err error) {

	destinationFileName := fs.getOriginFilename(origin)
	destinationPath := filepath.Dir(destinationFileName)

	err = fs.createPathIfNotExists(destinationPath)
	if err != nil {
		return err
	}

	temporaryFile, err := ioutil.TempFile("", "upload")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(temporaryFile.Name(), data, 0644)
	if err != nil {
		return err
	}

	defer func(name string) {
		_ = os.Remove(name)
	}(temporaryFile.Name())

	err = cp.CopyFile(temporaryFile.Name(), destinationFileName)
	if err != nil {
		return err
	}

	return
}

func (fs *fileStorage) Remove(origin *contracts.Origin) (err error) {
	originFileName := fs.getOriginFilename(origin)
	defer func(name string) {
		_ = os.Remove(name)
	}(originFileName)

	if origin.Type == "image" {
		miniature := contracts.Miniature{
			Type:     origin.Type,
			Category: origin.Category,
			Name:     origin.Name,
			Width:    0,
			Height:   0,
			Cast:     0,
		}
		thumbnailWildcard := fs.getImageThumbnailFilename(&miniature, "*", true)
		defer func(fs *fileStorage, wildcard string) {
			_ = fs.removeByWildcard(wildcard)
		}(fs, thumbnailWildcard)
	}

	return
}

func (fs *fileStorage) getFileInfo(fileName string) (info contracts.File, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return
	}

	stat, err := file.Stat()
	if err != nil {
		return
	}

	info.Size = stat.Size()
	info.ModificationTime = stat.ModTime()

	info.Contents, err = ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	info.ContentType = mime.TypeByExtension(filepath.Ext(fileName))
	if info.ContentType == "" {
		info.ContentType = http.DetectContentType(info.Contents)
	}

	return
}

func (fs *fileStorage) Read(origin *contracts.Origin) (info contracts.File, err error) {
	originFileName := fs.getOriginFilename(origin)
	info, err = fs.getFileInfo(originFileName)
	return
}

func (fs *fileStorage) ThumbnailExists(miniature *contracts.Miniature, extension string) bool {
	thumbnailFileName := fs.getImageThumbnailFilename(miniature, extension, false)
	_, err := os.Stat(thumbnailFileName)
	return !os.IsNotExist(err)
}

func (fs *fileStorage) ReadThumbnail(miniature *contracts.Miniature, extension string) (info contracts.File, err error) {
	thumbnailFileName := fs.getImageThumbnailFilename(miniature, extension, false)
	info, err = fs.getFileInfo(thumbnailFileName)
	return
}

func (fs *fileStorage) WriteThumbnail(miniature *contracts.Miniature, extension string, data []byte) (err error) {
	thumbnailFileName := fs.getImageThumbnailFilename(miniature, extension, false)
	thumbnailPath := filepath.Dir(thumbnailFileName)

	err = fs.createPathIfNotExists(thumbnailPath)
	if err != nil {
		return err
	}

	temporaryFile, err := ioutil.TempFile("", "upload")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(temporaryFile.Name(), data, 0644)
	if err != nil {
		return err
	}

	defer func(name string) {
		_ = os.Remove(name)
	}(temporaryFile.Name())

	err = cp.CopyFile(temporaryFile.Name(), thumbnailFileName)
	if err != nil {
		return err
	}

	return
}

func NewFileStorage(path string) Storage {
	result := &fileStorage{}
	result.SetStoragePath(path)
	return result
}
