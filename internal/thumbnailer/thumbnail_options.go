package thumbnailer

import "github.com/urvin/gokaru/internal/vips"

type RezizeMethod uint

const (
	ResizeMethodTensile RezizeMethod = iota
	ResizeMethodPrecize
	ResizeMethodInverse
)

type ThumbnailOptions struct {
	width                 uint
	height                uint
	imageType             vips.ImageType
	resizeMethod          RezizeMethod
	trim                  bool
	extent                bool
	opaqueBackground      bool
	transparentBackground bool
	padding               bool
}

func (to *ThumbnailOptions) Width() uint {
	return to.width
}

func (to *ThumbnailOptions) Height() uint {
	return to.height
}

func (to *ThumbnailOptions) ImageType() vips.ImageType {
	return to.imageType
}

func (to *ThumbnailOptions) Trim() bool {
	return to.trim
}

func (to *ThumbnailOptions) Extent() bool {
	return to.extent
}

func (to *ThumbnailOptions) OpaqueBackground() bool {
	return to.opaqueBackground
}

func (to *ThumbnailOptions) TransparentBackground() bool {
	return to.transparentBackground
}

func (to *ThumbnailOptions) Padding() bool {
	return to.padding
}

func (to *ThumbnailOptions) ResizeMethod() RezizeMethod {
	return to.resizeMethod
}

func (to *ThumbnailOptions) ResizeTensile() bool {
	return to.resizeMethod == ResizeMethodTensile
}

func (to *ThumbnailOptions) ResizePrecise() bool {
	return to.resizeMethod == ResizeMethodPrecize
}

func (to *ThumbnailOptions) ResizeInverse() bool {
	return to.resizeMethod == ResizeMethodInverse
}

func (to *ThumbnailOptions) SetWidth(width uint) {
	to.width = width
}

func (to *ThumbnailOptions) SetHeight(height uint) {
	to.height = height
}

func (to *ThumbnailOptions) SetImageTypeWithExtension(extension string) {
	to.imageType = vips.ImageTypeByExtension(extension)
}

func (to *ThumbnailOptions) SetOptionsWithCast(cast uint) {
	if to.hasCast(CAST_RESIZE_TENSILE, cast) {
		to.resizeMethod = ResizeMethodTensile
	}
	if to.hasCast(CAST_RESIZE_PRECISE, cast) {
		to.resizeMethod = ResizeMethodPrecize
	}
	if to.hasCast(CAST_RESIZE_INVERSE, cast) {
		to.resizeMethod = ResizeMethodInverse
	}

	to.trim = to.hasCast(CAST_TRIM, cast)
	to.extent = to.hasCast(CAST_EXTENT, cast)
	to.opaqueBackground = to.hasCast(CAST_OPAQUE_BACKGROUND, cast)
	to.transparentBackground = to.hasCast(CAST_TRANSPARENT_BACKGROUND, cast)
	to.padding = to.hasCast(CAST_TRIM_PADDING, cast)
}

func (to *ThumbnailOptions) hasCast(needle uint, haystack uint) bool {
	return (haystack & needle) > 0
}

func (to *ThumbnailOptions) SetTrim(trim bool) {
	to.trim = trim
}
