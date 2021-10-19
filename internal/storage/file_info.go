package storage

import (
	"io"
	"time"
)

type FileInfo struct {
	Size             int64
	ModificationTime time.Time
	ContentType      string
	Reader           io.Reader
}
