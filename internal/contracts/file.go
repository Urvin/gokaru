package contracts

import (
	"time"
)

type File struct {
	Size             int64
	ModificationTime time.Time
	ContentType      string
	Contents         []byte
}
