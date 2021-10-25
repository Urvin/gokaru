package thumbnailer

type Thumbnailer interface {
	Thumbnail(origin []byte, options ThumbnailOptions) (thumbnail []byte, later func([]byte) ([]byte, error), err error)
}
