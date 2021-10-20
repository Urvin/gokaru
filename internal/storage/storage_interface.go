package storage

import (
	"github.com/urvin/gokaru/internal/contracts"
	"io"
)

type Storage interface {
	Write(origin *contracts.Origin, data io.Reader) (err error)
	Remove(origin *contracts.Origin) (err error)
	Read(origin *contracts.Origin) (info contracts.File, err error)

	ThumbnailExists(miniature *contracts.Miniature, extension string) bool
	ReadThumbnail(miniature *contracts.Miniature, extension string) (info contracts.File, err error)
	WriteThumbnail(miniature *contracts.Miniature, extension string, data io.Reader) (err error)
}
