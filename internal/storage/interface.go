package storage

import "github.com/urvin/gokaru/internal/contracts"

type Storage interface {
	Write(origin *contracts.OriginDto, data []byte) (err error)

	Remove(origin *contracts.OriginDto) (err error)
	Read(origin *contracts.OriginDto) (info contracts.FileDto, err error)

	ThumbnailExists(miniature *contracts.MiniatureDto) bool
	ReadThumbnail(miniature *contracts.MiniatureDto) (info contracts.FileDto, err error)
	WriteThumbnail(miniature *contracts.MiniatureDto, data []byte) (err error)
}
