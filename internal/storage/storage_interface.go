package storage

import (
	"github.com/urvin/gokaru/internal/contracts"
)

type Storage interface {
	Write(origin *contracts.Origin, data []byte) (err error)

	Remove(origin *contracts.Origin) (err error)
	Read(origin *contracts.Origin) (info contracts.File, err error)

	ThumbnailExists(miniature *contracts.Miniature, extension string) bool
	ReadThumbnail(miniature *contracts.Miniature, extension string) (info contracts.File, err error)
	WriteThumbnail(miniature *contracts.Miniature, extension string, data []byte) (err error)
}
