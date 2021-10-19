package thumbnailer

import "io"

type Thumbnailer interface {
	Thumbnail(origin io.Reader, width, height, cast int, extension string) (thumbnail io.Reader, err error)
}
