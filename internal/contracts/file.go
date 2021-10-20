package contracts

import (
	"io"
	"time"
)

type File struct {
	Size             int64
	ModificationTime time.Time
	ContentType      string
	Reader           io.Reader
}
