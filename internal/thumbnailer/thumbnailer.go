package thumbnailer

import (
	"errors"
	"fmt"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"gopkg.in/alessio/shellescape.v1"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

type thumbnailer struct {
}

func (t *thumbnailer) Thumbnail(origin []byte, width, height, cast int, extension string) (thumbnail []byte, later func([]byte) ([]byte, error), err error) {
	originFile, err := ioutil.TempFile("", "thumbnail-or")
	if err != nil {
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(originFile.Name())

	err = ioutil.WriteFile(originFile.Name(), origin, 0644)
	if err != nil {
		return
	}

	destinationFile, err := ioutil.TempFile("", "thumbnail-dst.*."+extension)
	if err != nil {
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(destinationFile.Name())

	later, err = t.thumbnailFiles(originFile.Name(), destinationFile.Name(), width, height, cast, extension)
	if err != nil {
		return
	}

	thumbnail, err = ioutil.ReadFile(destinationFile.Name())
	if err != nil {
		return
	}

	return
}

func (t *thumbnailer) thumbnailFiles(origin, destination string, width, height, cast int, extension string) (later func([]byte) ([]byte, error), err error) {
	err = t.validateParams(width, height, cast, extension)
	if err != nil {
		return
	}

	destinationIsOpaque := helper.StringInSlice(extension, FORMATS_OPAQUE)
	opaqueBackgroundSet := false
	trimmed := false
	info, err := GetImageInfo(origin)
	if err != nil {
		return
	}

	imagickParams := []string{"-strip"}
	if info.Format == FORMAT_GIF {
		imagickParams = append(imagickParams, "-coalesce")
	}

	// Set transparent background
	if !info.Alpha &&
		t.hasCast(CAST_TRANSPARENT_BACKGROUND, cast) &&
		!t.hasCast(CAST_OPAQUE_BACKGROUND, cast) &&
		!destinationIsOpaque {
		firstPixelColor, er := t.getFirstPixelColor(origin)
		if er != nil {
			err = er
			return
		}
		if firstPixelColor != "" {
			imagickParams = append(imagickParams, "-alpha", "off", "-bordercolor", firstPixelColor, "-border", "1", "(", "+clone", "-fuzz", "20%", "-fill", "none", "-floodfill", "+0+0", firstPixelColor, "-alpha", "extract", "-geometry", "200%", "-blur", "0x0.5", "-morphology", "erode", "square:1", "-geometry", "50%", ")", "-compose", "CopyOpacity", "-composite", "-shave", "1", "-compose", "add")
		}
	}

	// Set opaque background
	if t.hasCast(CAST_OPAQUE_BACKGROUND, cast) || (info.Alpha && destinationIsOpaque) {
		imagickParams = append(imagickParams, "-background", "white", "-alpha", "remove", "-alpha", "off", "-fuzz", "5%", "-fill", "white", "-draw", "color 1,1 floodfill")
		opaqueBackgroundSet = true
	}

	// trim image
	if t.hasCast(CAST_TRIM, cast) {
		imagickParams = append(imagickParams, "-fuzz", "5%", "-trim")
		trimmed = true
	}

	// calculate resize
	forceExtent := false
	resizeFlag := "none"
	if width == 0 && height == 0 {
		width = info.Width
		height = info.Height
	} else if width == 0 {
		resizeFlag = "!"
		width = int(math.Floor(float64(info.Width) * float64(height) / float64(info.Height)))
	} else if height == 0 {
		resizeFlag = "!"
		height = int(math.Floor(float64(info.Height) * float64(width) / float64(info.Width)))
	} else if t.hasCast(CAST_RESIZE_TENSILE, cast) {
		resizeFlag = "!"
	} else if t.hasCast(CAST_RESIZE_PRECISE, cast) {
		resizeFlag = "^"
	} else if t.hasCast(CAST_RESIZE_INVERSE, cast) {
		resizeFlag = ""
	} else {
		forceExtent = true
	}

	resizeWidth := width
	resizeHeight := height
	padding := config.Get().Padding
	if t.hasCast(CAST_TRIM_PADDING, cast) &&
		t.hasCast(CAST_TRIM, cast) &&
		resizeWidth > 2*padding &&
		resizeHeight > 2*padding {
		resizeWidth -= 2 * padding
		resizeHeight -= 2 * padding
		forceExtent = true
	}

	if resizeFlag != "none" && (info.Width != resizeWidth || info.Height != resizeHeight || trimmed) {
		imagickParams = append(imagickParams, "-filter", RESIZE_FILTER, "-resize", strconv.Itoa(resizeWidth)+"x"+strconv.Itoa(resizeHeight)+resizeFlag)
	}

	// extent
	if forceExtent || t.hasCast(CAST_EXTENT, cast) {
		if !opaqueBackgroundSet {
			imagickParams = append(imagickParams, "-background", "none")
		}
		imagickParams = append(imagickParams, "-gravity", "center", "-extent", strconv.Itoa(width)+"x"+strconv.Itoa(height))
	}

	quality := t.getQuality(width, height, extension)

	// Animated image
	if info.Animated && helper.StringInSlice(extension, FORMATS_ANIMATED) {
		if extension == FORMAT_GIF {
			imagickParams = append(imagickParams, "+dither")

			paletteParams := append(imagickParams, "-colors", "256", "-unique-colors", "-depth 8")
			paletteFile, er := ioutil.TempFile("", "gifpalette")
			if er != nil {
				err = er
				return
			}
			defer func(name string) {
				_ = os.Remove(name)
			}(paletteFile.Name())

			err = t.execMagick(origin, FORMAT_GIF+":"+paletteFile.Name(), paletteParams)
			if err != nil {
				return
			}

			imagickParams = append(imagickParams, "+dither", "-remap", paletteFile.Name(), "-layers", "optimize")
			err = t.execMagick(origin, destination, imagickParams)
			if err != nil {
				return
			}
		} else {
			imagickParams = append(imagickParams, "-layers", "optimize", "-quality", strconv.Itoa(quality.Quality))
			err = t.execMagick(origin, destination, imagickParams)
		}
	} else {
		if info.Animated {
			imagickParams = append(imagickParams, "-flatten")
			origin += "[0]"
		}

		if extension == FORMAT_JPG {
			destinationTempFile, er := ioutil.TempFile("", "thumbnail-jp")
			if er != nil {
				err = er
				return
			}
			defer func(name string) {
				_ = os.Remove(name)
			}(destinationTempFile.Name())

			err = t.execMagick(origin, FORMAT_PNG+":"+destinationTempFile.Name(), append(imagickParams, "-quality", "100"))
			if err != nil {
				return
			}

			_, err = helper.Exec("sh", "-c", "mozjpeg -optimize -progressive -quality "+strconv.Itoa(quality.Quality)+" "+shellescape.Quote(destinationTempFile.Name())+" > "+shellescape.Quote(destination))
			if err != nil {
				return
			}

		} else {
			err = t.execMagick(origin, destination, append(imagickParams, "-quality", strconv.Itoa(quality.Quality)))
			if err != nil {
				return
			}

			if extension == FORMAT_PNG {
				_, err = helper.Exec("pngquant", "--force", "--quality="+strconv.Itoa(quality.QualityMin)+"-"+strconv.Itoa(quality.Quality), "--strip", "--skip-if-larger", "--speed", "1", "--ext", ".png", "--verbose", "--", destination)
				if err != nil {
					return
				}
				later = t.laterOptimizePng
			}
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

	info, err := GetImageInfo(uncompressedFile.Name())
	if err != nil {
		return
	}

	quality := t.getQuality(info.Width, info.Height, info.Format)

	_, err = helper.Exec("zopflipng", "--iterations="+strconv.Itoa(quality.Iterations), "-y", "--filters=01234mepb", "--lossy_8bit", "--lossy_transparent", uncompressedFile.Name(), compressedFile.Name())
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

func (t *thumbnailer) hasCast(needle int, haystack int) bool {
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

func (t *thumbnailer) getQuality(width, height int, format string) Quality {
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
