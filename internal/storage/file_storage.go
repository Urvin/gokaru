package storage

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"github.com/urvin/gokaru/internal/helper"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const IMAGE_ORIGIN_PATH = "origin"
const IMAGE_THUMBNAIL_PATH = "thumbnail"
const SNIFF_LENGTH = 512

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

func (fs *fileStorage) getOriginFilename(sourceType, fileCategory, fileName string) string {
	hashedFileName := fs.hashFileName(fileName)
	hashedFilePath := hashedFileName[0:2] + "/" + hashedFileName[2:4]

	result := fs.storagePath + "/" + sourceType
	if sourceType == "image" {
		result = result + "/" + IMAGE_ORIGIN_PATH
	}
	result = result + "/" + fileCategory + "/" + hashedFilePath + "/" + hashedFileName

	return result
}

func (fs *fileStorage) getImageThumbnailFilename(sourceType, fileCategory, fileName string, width, height, cast int, extension string, del bool) string {
	hashedFileName := fs.hashFileName(fileName)
	hashedFilePath := hashedFileName[0:2] + "/" + hashedFileName[2:4]

	castPath := "*"
	extensionPart := "*"
	if !del {
		castPath = strconv.Itoa(width) + "x" + strconv.Itoa(height) + "x" + strconv.Itoa(cast)
		extensionPart = extension
	}

	result := fs.storagePath + "/" + sourceType + "/" + IMAGE_THUMBNAIL_PATH
	result = result + "/" + fileCategory + "/" + castPath + "/" + hashedFilePath + "/" + hashedFileName + "." + extensionPart

	return result
}

func (fs *fileStorage) SetStoragePath(path string) {
	fs.storagePath, _ = filepath.Abs(path)
}

func (fs *fileStorage) Write(sourceType, fileCategory, fileName string, data io.Reader) (err error) {

	destinationFileName := fs.getOriginFilename(sourceType, fileCategory, fileName)
	destinationPath := filepath.Dir(destinationFileName)

	err = fs.createPathIfNotExists(destinationPath)
	if err != nil {
		return err
	}

	temporaryFile, err := ioutil.TempFile("", "upload")
	if err != nil {
		return err
	}
	_, err = io.Copy(temporaryFile, data)
	if err != nil {
		return err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(temporaryFile.Name())

	err = helper.CopyFile(temporaryFile.Name(), destinationFileName)
	if err != nil {
		return err
	}

	return
}

func (fs *fileStorage) Remove(sourceType, fileCategory, fileName string) (err error) {
	originFileName := fs.getOriginFilename(sourceType, fileCategory, fileName)
	defer func(name string) {
		_ = os.Remove(name)
	}(originFileName)

	if sourceType == "image" {
		thumbnailWildcard := fs.getImageThumbnailFilename(sourceType, fileCategory, fileName, 0, 0, 0, "*", true)
		defer func(fs *fileStorage, wildcard string) {
			_ = fs.removeByWildcard(wildcard)
		}(fs, thumbnailWildcard)
	}

	return
}

func (fs *fileStorage) getFileInfo(fileName string) (info FileInfo, err error) {
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

	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		// read a chunk to decide between utf-8 text and binary
		var buf [SNIFF_LENGTH]byte
		n, _ := io.ReadFull(file, buf[:])
		contentType = http.DetectContentType(buf[:n])
		_, err = file.Seek(0, io.SeekStart) // rewind to output whole file
		if err != nil {
			return
		}
	}
	info.ContentType = contentType

	info.Reader = bufio.NewReader(file)
	return
}

func (fs *fileStorage) Read(sourceType, fileCategory, fileName string) (info FileInfo, err error) {
	originFileName := fs.getOriginFilename(sourceType, fileCategory, fileName)
	info, err = fs.getFileInfo(originFileName)
	return
}

func (fs *fileStorage) ThumbnailExists(sourceType, fileCategory, fileName string, width, height, cast int, extension string) bool {
	thumbnailFileName := fs.getImageThumbnailFilename(sourceType, fileCategory, fileName, width, height, cast, extension, false)
	_, err := os.Stat(thumbnailFileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func (fs *fileStorage) ReadThumbnail(sourceType, fileCategory, fileName string, width, height, cast int, extension string) (info FileInfo, err error) {
	thumbnailFileName := fs.getImageThumbnailFilename(sourceType, fileCategory, fileName, width, height, cast, extension, false)
	info, err = fs.getFileInfo(thumbnailFileName)
	return
}

func (fs *fileStorage) WriteThumbnail(sourceType, fileCategory, fileName string, width, height, cast int, extension string, data io.Reader) (err error) {
	thumbnailFileName := fs.getImageThumbnailFilename(sourceType, fileCategory, fileName, width, height, cast, extension, false)
	thumbnailPath := filepath.Dir(thumbnailFileName)

	err = fs.createPathIfNotExists(thumbnailPath)
	if err != nil {
		return err
	}

	temporaryFile, err := ioutil.TempFile("", "upload")
	if err != nil {
		return err
	}

	_, err = io.Copy(temporaryFile, data)
	if err != nil {
		return err
	}

	defer func(name string) {
		_ = os.Remove(name)
	}(temporaryFile.Name())

	err = helper.CopyFile(temporaryFile.Name(), thumbnailFileName)
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
