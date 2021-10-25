package vips

/*
#include "vips.h"
*/
import "C"

type ImageType int

const (
	ImageTypeUnknown = ImageType(C.UNKNOWN)
	ImageTypeJPEG    = ImageType(C.JPEG)
	ImageTypePNG     = ImageType(C.PNG)
	ImageTypeWEBP    = ImageType(C.WEBP)
	ImageTypeGIF     = ImageType(C.GIF)
	ImageTypeICO     = ImageType(C.ICO)
	ImageTypeSVG     = ImageType(C.SVG)
	ImageTypeHEIC    = ImageType(C.HEIC)
	ImageTypeAVIF    = ImageType(C.AVIF)
	ImageTypeBMP     = ImageType(C.BMP)
	ImageTypeTIFF    = ImageType(C.TIFF)
)

var (
	ImageTypes = map[string]ImageType{
		"jpg":  ImageTypeJPEG,
		"jpeg": ImageTypeJPEG,
		"png":  ImageTypePNG,
		"webp": ImageTypeWEBP,
		"gif":  ImageTypeGIF,
		"ico":  ImageTypeICO,
		"svg":  ImageTypeSVG,
		"heic": ImageTypeHEIC,
		"avif": ImageTypeAVIF,
		"bmp":  ImageTypeBMP,
		"tiff": ImageTypeTIFF,
	}

	mimes = map[ImageType]string{
		ImageTypeJPEG: "image/jpeg",
		ImageTypePNG:  "image/png",
		ImageTypeWEBP: "image/webp",
		ImageTypeGIF:  "image/gif",
		ImageTypeICO:  "image/x-icon",
		ImageTypeSVG:  "image/svg+xml",
		ImageTypeHEIC: "image/heif",
		ImageTypeAVIF: "image/avif",
		ImageTypeBMP:  "image/bmp",
		ImageTypeTIFF: "image/tiff",
	}
)

func (it ImageType) String() string {
	for k, v := range ImageTypes {
		if v == it {
			return k
		}
	}
	return ""
}

func (it ImageType) Mime() string {
	if mime, ok := mimes[it]; ok {
		return mime
	}

	return "application/octet-stream"
}

func (it ImageType) SupportsAlpha() bool {
	return it != ImageTypeJPEG && it != ImageTypeBMP
}

func (it ImageType) SupportsColourProfile() bool {
	return it == ImageTypeJPEG ||
		it == ImageTypePNG ||
		it == ImageTypeWEBP ||
		it == ImageTypeAVIF
}

func (it ImageType) SupportsAnimation() bool {
	return it == ImageTypeGIF || (it == ImageTypeWEBP && SupportsWebpAnimation())
}

func ImageTypeByByMime(mime string) ImageType {
	for key, value := range mimes {
		if value == mime {
			return key
		}
	}
	return ImageTypeUnknown
}

func ImageTypeByExtension(extension string) ImageType {
	if it, ok := ImageTypes[extension]; ok {
		return it
	}
	return ImageTypeUnknown
}
