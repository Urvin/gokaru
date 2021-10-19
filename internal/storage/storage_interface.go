package storage

import "io"

type Storage interface {
	Write(sourceType, fileCategory, fileName string, data io.Reader) (err error)
	Remove(sourceType, fileCategory, fileName string) (err error)
	Read(sourceType, fileCategory, fileName string) (info FileInfo, err error)

	ThumbnailExists(sourceType, fileCategory, fileName string, width, height, cast int, extension string) bool
	ReadThumbnail(sourceType, fileCategory, fileName string, width, height, cast int, extension string) (info FileInfo, err error)
	WriteThumbnail(sourceType, fileCategory, fileName string, width, height, cast int, extension string, data io.Reader) (err error)
}
