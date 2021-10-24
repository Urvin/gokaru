package thumbnailer

import (
	"errors"
	"fmt"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"gopkg.in/gographics/imagick.v3/imagick"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

type thumbnailer struct {
}

func (t *thumbnailer) Thumbnail(origin []byte, width, height, cast int, extension string) (thumbnail []byte, later func([]byte) ([]byte, error), err error) {
	err = t.validateParams(width, height, cast, extension)
	if err != nil {
		return
	}
	thumbnail, later, err = t.thumbnailBytes(origin, uint(width), uint(height), uint(cast), extension)
	return
}

func (t *thumbnailer) thumbnailBytes(origin []byte, width, height, cast uint, extension string) (thumbnail []byte, later func([]byte) ([]byte, error), err error) {

	destinationIsOpaque := helper.StringInSlice(extension, FORMATS_OPAQUE)
	opaqueBackgroundSet := false
	trimmed := false

	_, depth := imagick.GetQuantumDepth()
	fuzz5prc := math.Pow(2, float64(depth)) * .05

	mw := imagick.NewMagickWand()
	err = mw.ReadImageBlob(origin)
	if err != nil {
		return
	}

	info, err := GetImageInfoWand(mw)
	if err != nil {
		return
	}

	err = mw.StripImage()
	if err != nil {
	}
	if info.Format == FORMAT_GIF {
		mw.CoalesceImages()
	}

	// Set transparent background
	if !info.Alpha &&
		t.hasCast(CAST_TRANSPARENT_BACKGROUND, cast) &&
		!t.hasCast(CAST_OPAQUE_BACKGROUND, cast) &&
		!destinationIsOpaque {

		//firstPixel, er := mw.GetImagePixelColor(0, 0)
		//if er != nil {
		//	err = er
		//	return
		//}
		// TODO: прозрачный фон
	}

	// Set opaque background
	if t.hasCast(CAST_OPAQUE_BACKGROUND, cast) || (info.Alpha && destinationIsOpaque) {
		bgcolor := imagick.NewPixelWand()
		bgcolor.SetColor("white")

		err = mw.SetImageBackgroundColor(bgcolor)
		if err != nil {
			return
		}

		err = mw.SetImageAlpha(1)
		if err != nil {
			return
		}
		err = mw.FloodfillPaintImage(bgcolor, fuzz5prc, bgcolor, 0, 0, false)
		if err != nil {
			return
		}

		opaqueBackgroundSet = true
	}

	// trim image
	if t.hasCast(CAST_TRIM, cast) {
		err = mw.TrimImage(fuzz5prc)
		if err != nil {
			return
		}
		trimmed = true
	}

	forceExtent := false
	shouldResize := false

	if width == 0 && height == 0 {
		width = info.Width
		height = info.Height
	} else if width == 0 {
		shouldResize = true
		width = uint(math.Floor(float64(info.Width) * float64(height) / float64(info.Height)))
	} else if height == 0 {
		shouldResize = true
		height = uint(math.Floor(float64(info.Height) * float64(width) / float64(info.Width)))
	} else if t.hasCast(CAST_RESIZE_TENSILE, cast) {
		shouldResize = true
	} else if t.hasCast(CAST_RESIZE_PRECISE, cast) {
		shouldResize = true
		width, height, err = calculateWHWithAspectRatio(info.Width, info.Height, width, height, true)
		if err != nil {
			return
		}
	} else if t.hasCast(CAST_RESIZE_INVERSE, cast) {
		shouldResize = true
		width, height, err = calculateWHWithAspectRatio(info.Width, info.Height, width, height, false)
		if err != nil {
			return
		}
	} else {
		forceExtent = true
	}

	resizeWidth := width
	resizeHeight := height
	padding := uint(config.Get().Padding)
	if t.hasCast(CAST_TRIM_PADDING, cast) && t.hasCast(CAST_TRIM, cast) &&
		resizeWidth > 2*padding && resizeHeight > 2*padding {
		resizeWidth -= 2 * padding
		resizeHeight -= 2 * padding
		forceExtent = true
	}

	if shouldResize && (info.Width != resizeWidth || info.Height != resizeHeight || trimmed) {
		err = mw.ResizeImage(resizeWidth, resizeHeight, RESIZE_FILTER)
		if err != nil {
			return
		}
	}

	// extent
	if forceExtent || t.hasCast(CAST_EXTENT, cast) {
		if !opaqueBackgroundSet {
			bgnone := imagick.NewPixelWand()
			bgnone.SetColor("none")
			err = mw.SetImageBackgroundColor(bgnone)
			if err != nil {
				return
			}
		}
		err = mw.SetGravity(imagick.GRAVITY_CENTER)
		if err != nil {
			return
		}
	}

	quality := t.getQuality(width, height, extension)

	if info.Animated && helper.StringInSlice(extension, FORMATS_ANIMATED) {
		if extension == FORMAT_GIF {
			//TODO: optimize gif
		} else {
			mw.OptimizeImageLayers()
			err = mw.SetImageCompressionQuality(quality.Quality)
			if err != nil {
				return
			}
		}
	} else {
		if info.Animated {
			mw.MergeImageLayers(imagick.IMAGE_LAYER_FLATTEN)
		}

		imQuality := quality.Quality
		imExtension := extension

		if extension == FORMAT_JPG || extension == FORMAT_PNG {
			imQuality = 100
		}

		err = mw.SetFormat(imExtension)
		if err != nil {
			return
		}

		err = mw.SetImageCompressionQuality(imQuality)
		if err != nil {
			return
		}

		thumbnail = mw.GetImageBlob()

		if extension == FORMAT_JPG {
			thumbnail, err = mozjpeg(thumbnail, quality.Quality)
			return
		} else if extension == FORMAT_PNG {
			thumbnail, err = pngquant(thumbnail, quality.QualityMin, quality.Quality)
			if err != nil {
				later = t.laterOptimizePng
			}
			return
		}

	}

	return
}

func (t *thumbnailer) laterOptimizePng(uncompressed []byte) (compressed []byte, err error) {
	uncompressedFile, err := ioutil.TempFile("", "thumbnail-pngi")
	defer func(name string) {
		_ = os.Remove(name)
	}(uncompressedFile.Name())

	err = ioutil.WriteFile(uncompressedFile.Name(), uncompressed, 0644)
	if err != nil {
		return
	}

	compressedFile, err := ioutil.TempFile("", "thumbnail-pngo")
	defer func(name string) {
		_ = os.Remove(name)
	}(compressedFile.Name())

	info, err := GetImageInfoFile(uncompressedFile.Name())
	if err != nil {
		return
	}

	quality := t.getQuality(info.Width, info.Height, info.Format)

	_, err = helper.Exec("zopflipng", "--iterations="+strconv.FormatUint(uint64(quality.Iterations), 10), "-y", "--filters=01234mepb", "--lossy_8bit", "--lossy_transparent", uncompressedFile.Name(), compressedFile.Name())
	if err != nil {
		return
	}

	compressed, err = ioutil.ReadFile(compressedFile.Name())
	if err != nil {
		return
	}

	return
}

func (t *thumbnailer) validateParams(width, height, cast int, extension string) (err error) {
	if width < 0 {
		err = errors.New("width should not be less than 0")
		return
	}
	if height < 0 {
		err = errors.New("height should not be less than 0")
		return
	}
	if cast < 0 {
		err = errors.New("cast should not be less than 0")
		return
	}
	if !helper.StringInSlice(extension, FORMATS_ALLOWED) {
		err = fmt.Errorf("'%s' in not a valid thumbnail format", extension)
	}
	return
}

func (t *thumbnailer) hasCast(needle uint, haystack uint) bool {
	return (haystack & needle) > 0
}

func (t *thumbnailer) getFirstPixelColor(fileName string) (color string, err error) {

	output, er := helper.Exec("convert", fileName, "-format", "%[pixel:p{0,0}]", "info:-")
	if er != nil {
		err = er
		return
	}
	color = strings.Trim(output, "\n\r\t ")
	return
}

func (t *thumbnailer) getQuality(width, height uint, format string) Quality {
	defaultQuality := Quality{
		Quality:    config.Get().QualityDefault,
		QualityMin: config.Get().QualityDefault,
		Iterations: 100,
	}

	result := defaultQuality
	halfPerimeter := width + height

	for _, qualityFormat := range config.Get().Quality {
		if qualityFormat.Format == format {
			result.Quality = qualityFormat.Quality
			result.QualityMin = qualityFormat.QualityMin
			result.Iterations = qualityFormat.Iterations

			for _, condition := range qualityFormat.Conditions {
				if halfPerimeter >= condition.From && halfPerimeter < condition.To {
					result.Quality = condition.Quality
					result.QualityMin = condition.QualityMin
					result.Iterations = condition.Iterations
					break
				}
			}
			break
		}
	}

	if result.Quality <= 0 {
		result = defaultQuality
	}
	if result.Quality > 100 {
		result.Quality = 100
	}
	if result.QualityMin <= 0 {
		result.QualityMin = result.Quality
	}
	if result.QualityMin > 100 {
		result.QualityMin = 100
	}
	if result.QualityMin > result.Quality {
		result.QualityMin = result.Quality
	}
	if result.Iterations <= 0 {
		result.Iterations = 100
	}

	return result
}

func (t *thumbnailer) execMagick(origin, destination string, params []string) (err error) {
	params = append([]string{origin}, append(params, destination)...)
	_, err = helper.Exec("convert", params...)
	if err != nil {
		return
	}
	return
}

func NewThumbnailer() Thumbnailer {
	result := &thumbnailer{}
	return result
}
