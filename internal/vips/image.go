package vips

/*
#cgo pkg-config: vips
#include "vips.h"
*/
import "C"
import (
	"math"
	"unsafe"
)

type Image struct {
	VipsImage *C.VipsImage
}

func (img *Image) Width() int {
	return int(img.VipsImage.Xsize)
}

func (img *Image) Height() int {
	return int(img.VipsImage.Ysize)
}

func (img *Image) Load(data []byte, imgtype ImageType, shrink int, scale float64, pages int) error {
	var tmp *C.VipsImage

	err := C.int(0)

	switch imgtype {
	case ImageTypeJPEG:
		err = C.vips_jpegload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(shrink), &tmp)
	case ImageTypePNG:
		err = C.vips_pngload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	case ImageTypeWEBP:
		err = C.vips_webpload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.double(scale), C.int(pages), &tmp)
	case ImageTypeGIF:
		err = C.vips_gifload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.int(pages), &tmp)
	case ImageTypeSVG:
		err = C.vips_svgload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), C.double(scale), &tmp)
	case ImageTypeHEIC, ImageTypeAVIF:
		err = C.vips_heifload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	case ImageTypeBMP:
		err = C.vips_bmpload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	case ImageTypeTIFF:
		err = C.vips_tiffload_go(unsafe.Pointer(&data[0]), C.size_t(len(data)), &tmp)
	}
	if err != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) SaveJpeg(options JpegSaveOptions) ([]byte, error) {
	var ptr unsafe.Pointer
	err := C.int(0)
	imgsize := C.size_t(0)

	err = C.vips_savejpeg_ext_go(img.VipsImage, &ptr, &imgsize,
		C.int(options.Quality),
		gbool(options.Interlace),
		gbool(options.OptimizeCoding),
		C.int(options.SubsampleMode),
		gbool(options.TrellisQuant),
		gbool(options.OvershootDeringing),
		gbool(options.OptimizeScans),
		C.int(options.QuantTable),
	)

	if err != 0 {
		C.g_free_go(&ptr)
		return nil, vipsError()
	}
	b := ptrToBytes(ptr, int(imgsize))
	return b, nil
}

func (img *Image) SavePng(options PngSaveOptions) ([]byte, error) {
	var ptr unsafe.Pointer
	err := C.int(0)
	imgsize := C.size_t(0)

	err = C.vips_pngsave_go(img.VipsImage, &ptr, &imgsize,
		gbool(options.Interlace),
		ibool(options.Quantize),
		C.int(options.Colors),
		C.int(options.Compression))

	if err != 0 {
		C.g_free_go(&ptr)
		return nil, vipsError()
	}
	b := ptrToBytes(ptr, int(imgsize))
	return b, nil
}

func (img *Image) SaveWebp(options WebpSaveOptions) ([]byte, error) {
	var ptr unsafe.Pointer
	err := C.int(0)
	imgsize := C.size_t(0)

	err = C.vips_webpsave_go(img.VipsImage, &ptr, &imgsize,
		C.int(options.Quality),
		gbool(options.Lossless),
		gbool(options.SmartSubsample),
		C.int(options.ReductionEffort))

	if err != 0 {
		C.g_free_go(&ptr)
		return nil, vipsError()
	}
	b := ptrToBytes(ptr, int(imgsize))
	return b, nil
}

func (img *Image) SaveAvif(options AvifSaveOptions) ([]byte, error) {
	var ptr unsafe.Pointer
	err := C.int(0)
	imgsize := C.size_t(0)

	err = C.vips_avifsave_go(img.VipsImage, &ptr, &imgsize,
		C.int(options.Quality),
		C.int(options.Speed),
		gbool(options.Lossless))

	if err != 0 {
		C.g_free_go(&ptr)
		return nil, vipsError()
	}
	b := ptrToBytes(ptr, int(imgsize))
	return b, nil
}

func (img *Image) Save(imgtype ImageType, quality int) ([]byte, error) {
	var ptr unsafe.Pointer

	err := C.int(0)
	imgsize := C.size_t(0)

	switch imgtype {
	case ImageTypeJPEG:
		err = C.vips_jpegsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality), C.int(1))
	case ImageTypePNG:
		err = C.vips_pngsave_go(img.VipsImage, &ptr, &imgsize, gbool(false), gbool(false), C.int(255), C.int(6))
	case ImageTypeWEBP:
		err = C.vips_webpsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality), gbool(false), gbool(false), C.int(4))
	case ImageTypeGIF:
		err = C.vips_gifsave_go(img.VipsImage, &ptr, &imgsize)
	case ImageTypeAVIF:
		err = C.vips_avifsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality), C.int(5), gbool(false))
	case ImageTypeBMP:
		err = C.vips_bmpsave_go(img.VipsImage, &ptr, &imgsize)
	case ImageTypeTIFF:
		err = C.vips_tiffsave_go(img.VipsImage, &ptr, &imgsize, C.int(quality))
	}
	if err != 0 {
		C.g_free_go(&ptr)
		return nil, vipsError()
	}

	b := ptrToBytes(ptr, int(imgsize))

	return b, nil
}

func (img *Image) Clear() {
	if img.VipsImage != nil {
		C.clear_image(&img.VipsImage)
	}
}

func (img *Image) Arrayjoin(in []*Image) error {
	var tmp *C.VipsImage

	arr := make([]*C.VipsImage, len(in))
	for i, im := range in {
		arr[i] = im.VipsImage
	}

	if C.vips_arrayjoin_go(&arr[0], &tmp, C.int(len(arr))) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) IsAnimated() bool {
	return C.vips_is_animated(img.VipsImage) > 0
}

func (img *Image) HasAlpha() bool {
	return C.vips_image_hasalpha_go(img.VipsImage) > 0
}

func (img *Image) GetInt(name string) (int, error) {
	var i C.int

	if C.vips_image_get_int(img.VipsImage, cachedCString(name), &i) != 0 {
		return 0, vipsError()
	}
	return int(i), nil
}

func (img *Image) GetIntDefault(name string, def int) (int, error) {
	if C.vips_image_get_typeof(img.VipsImage, cachedCString(name)) == 0 {
		return def, nil
	}

	return img.GetInt(name)
}

func (img *Image) GetIntSlice(name string) ([]int, error) {
	var ptr unsafe.Pointer
	size := C.int(0)

	if C.vips_image_get_array_int_go(img.VipsImage, cachedCString(name), (**C.int)(unsafe.Pointer(&ptr)), &size) != 0 {
		return nil, vipsError()
	}

	if size == 0 {
		return []int{}, nil
	}

	cOut := (*[math.MaxInt32]C.int)(ptr)[:int(size):int(size)]
	out := make([]int, int(size))

	for i, el := range cOut {
		out[i] = int(el)
	}

	return out, nil
}

func (img *Image) GetIntSliceDefault(name string, def []int) ([]int, error) {
	if C.vips_image_get_typeof(img.VipsImage, cachedCString(name)) == 0 {
		return def, nil
	}

	return img.GetIntSlice(name)
}

func (img *Image) SetInt(name string, value int) {
	C.vips_image_set_int(img.VipsImage, cachedCString(name), C.int(value))
}

func (img *Image) SetIntSlice(name string, value []int) {
	in := make([]C.int, len(value))
	for i, el := range value {
		in[i] = C.int(el)
	}
	C.vips_image_set_array_int_go(img.VipsImage, cachedCString(name), &in[0], C.int(len(value)))
}

func (img *Image) CastUchar() error {
	var tmp *C.VipsImage

	if C.vips_image_get_format(img.VipsImage) != C.VIPS_FORMAT_UCHAR {
		if C.vips_cast_go(img.VipsImage, &tmp, C.VIPS_FORMAT_UCHAR) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *Image) Rad2Float() error {
	var tmp *C.VipsImage

	if C.vips_image_get_coding(img.VipsImage) == C.VIPS_CODING_RAD {
		if C.vips_rad2float_go(img.VipsImage, &tmp) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *Image) Resize(scale float64, hasAlpa bool) error {
	var tmp *C.VipsImage

	if hasAlpa {
		if C.vips_resize_with_premultiply(img.VipsImage, &tmp, C.double(scale)) != 0 {
			return vipsError()
		}
	} else {
		if C.vips_resize_go(img.VipsImage, &tmp, C.double(scale)) != 0 {
			return vipsError()
		}
	}

	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Orientation() C.int {
	return C.vips_get_orientation(img.VipsImage)
}

func (img *Image) Rotate(angle int) error {
	var tmp *C.VipsImage

	vipsAngle := (angle / 90) % 4

	if C.vips_rot_go(img.VipsImage, &tmp, C.VipsAngle(vipsAngle)) != 0 {
		return vipsError()
	}

	C.vips_autorot_remove_angle(tmp)

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Flip() error {
	var tmp *C.VipsImage

	if C.vips_flip_horizontal_go(img.VipsImage, &tmp) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Crop(left, top, width, height int) error {
	var tmp *C.VipsImage

	if C.vips_extract_area_go(img.VipsImage, &tmp, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Extract(out *Image, left, top, width, height int) error {
	if C.vips_extract_area_go(img.VipsImage, &out.VipsImage, C.int(left), C.int(top), C.int(width), C.int(height)) != 0 {
		return vipsError()
	}
	return nil
}

func (img *Image) SmartCrop(width, height int) error {
	var tmp *C.VipsImage

	if C.vips_smartcrop_go(img.VipsImage, &tmp, C.int(width), C.int(height)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Trim(threshold float64, smart bool, color RgbColor, equalHor bool, equalVer bool) error {
	var tmp *C.VipsImage

	if err := img.CopyMemory(); err != nil {
		return err
	}

	if C.vips_trim(img.VipsImage, &tmp, C.double(threshold),
		gbool(smart), C.double(color.R), C.double(color.G), C.double(color.B),
		gbool(equalHor), gbool(equalVer)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) EnsureAlpha() error {
	var tmp *C.VipsImage

	if C.vips_ensure_alpha(img.VipsImage, &tmp) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Flatten(bg RgbColor) error {
	var tmp *C.VipsImage

	if C.vips_flatten_go(img.VipsImage, &tmp, C.double(bg.R), C.double(bg.G), C.double(bg.B)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Blur(sigma float32) error {
	var tmp *C.VipsImage

	if C.vips_gaussblur_go(img.VipsImage, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Sharpen(sigma float32) error {
	var tmp *C.VipsImage

	if C.vips_sharpen_go(img.VipsImage, &tmp, C.double(sigma)) != 0 {
		return vipsError()
	}

	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) ImportColourProfile() error {
	var tmp *C.VipsImage

	if img.VipsImage.Coding != C.VIPS_CODING_NONE {
		return nil
	}

	if img.VipsImage.BandFmt != C.VIPS_FORMAT_UCHAR && img.VipsImage.BandFmt != C.VIPS_FORMAT_USHORT {
		return nil
	}

	if C.vips_has_embedded_icc(img.VipsImage) == 0 {
		return nil
	}

	if C.vips_icc_import_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		//logWarning("Can't import ICC profile: %s", vipsError())
	}

	return nil
}

func (img *Image) ExportColourProfile() error {
	var tmp *C.VipsImage

	// Don't export is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_export_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		//logWarning("Can't export ICC profile: %s", vipsError())
	}

	return nil
}

func (img *Image) ExportColourProfileToSRGB() error {
	var tmp *C.VipsImage

	// Don't export is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_export_srgb(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		//logWarning("Can't export ICC profile: %s", vipsError())
	}

	return nil
}

func (img *Image) TransformColourProfile() error {
	var tmp *C.VipsImage

	// Don't transform is there's no embedded profile or embedded profile is sRGB
	if C.vips_has_embedded_icc(img.VipsImage) == 0 || C.vips_icc_is_srgb_iec61966(img.VipsImage) == 1 {
		return nil
	}

	if C.vips_icc_transform_go(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		//logWarning("Can't transform ICC profile: %s", vipsError())
	}

	return nil
}

func (img *Image) RemoveColourProfile() error {
	var tmp *C.VipsImage

	if C.vips_icc_remove(img.VipsImage, &tmp) == 0 {
		C.swap_and_clear(&img.VipsImage, tmp)
	} else {
		//logWarning("Can't remove ICC profile: %s", vipsError())
	}

	return nil
}

func (img *Image) LinearColourspace() error {
	return img.Colorspace(C.VIPS_INTERPRETATION_scRGB)
}

func (img *Image) RgbColourspace() error {
	return img.Colorspace(C.VIPS_INTERPRETATION_sRGB)
}

func (img *Image) Colorspace(colorspace C.VipsInterpretation) error {
	if img.VipsImage.Type != colorspace {
		var tmp *C.VipsImage

		if C.vips_colourspace_go(img.VipsImage, &tmp, colorspace) != 0 {
			return vipsError()
		}
		C.swap_and_clear(&img.VipsImage, tmp)
	}

	return nil
}

func (img *Image) CopyMemory() error {
	var tmp *C.VipsImage
	if tmp = C.vips_image_copy_memory(img.VipsImage); tmp == nil {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}

func (img *Image) Replicate(width, height int) error {
	var tmp *C.VipsImage

	if C.vips_replicate_go(img.VipsImage, &tmp, C.int(width), C.int(height)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Embed(width, height int, offX, offY int, bg RgbColor, transpBg bool) error {
	var tmp *C.VipsImage

	if err := img.RgbColourspace(); err != nil {
		return err
	}

	var bgc []C.double
	if transpBg {
		if !img.HasAlpha() {
			if C.vips_addalpha_go(img.VipsImage, &tmp) != 0 {
				return vipsError()
			}
			C.swap_and_clear(&img.VipsImage, tmp)
		}

		bgc = []C.double{C.double(0)}
	} else {
		bgc = []C.double{C.double(bg.R), C.double(bg.G), C.double(bg.B), 1.0}
	}

	bgn := minInt(int(img.VipsImage.Bands), len(bgc))

	if C.vips_embed_go(img.VipsImage, &tmp, C.int(offX), C.int(offY), C.int(width), C.int(height), &bgc[0], C.int(bgn)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) ApplyWatermark(wm *Image, opacity float64) error {
	var tmp *C.VipsImage

	if C.vips_apply_watermark(img.VipsImage, wm.VipsImage, &tmp, C.double(opacity)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Strip() error {
	var tmp *C.VipsImage

	if C.vips_strip(img.VipsImage, &tmp) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)

	return nil
}

func (img *Image) Thumbnail(width, height int) error {
	var tmp *C.VipsImage

	if C.vips_thumbnail_go(img.VipsImage, &tmp, C.int(width), C.int(height)) != 0 {
		return vipsError()
	}
	C.swap_and_clear(&img.VipsImage, tmp)
	return nil
}
