package thumbnailer

type Thumbnailer interface {
	Thumbnail(origin []byte, width, height, cast int, extension string) (thumbnail []byte, later func([]byte) ([]byte, error), err error)
}
