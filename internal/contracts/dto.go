package contracts

import (
	"strconv"
	"time"
)

type OriginDto struct {
	Type     string
	Category string
	Name     string
}

type FileDto struct {
	Size             int64
	ModificationTime time.Time
	ContentType      string
	Contents         []byte
}

type MiniatureDto struct {
	Type      string
	Category  string
	Name      string
	Extension string
	Width     int
	Height    int
	Cast      int
}

func (miniature *MiniatureDto) Hash() string {
	return miniature.Type + "/" +
		miniature.Category + "/" +
		miniature.Name + "." + miniature.Extension + "/" +
		strconv.Itoa(miniature.Width) + "/" +
		strconv.Itoa(miniature.Height) + "/" +
		strconv.Itoa(miniature.Cast)
}
