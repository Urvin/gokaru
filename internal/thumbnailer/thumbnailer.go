package thumbnailer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"gopkg.in/alessio/shellescape.v1"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

type thumbnailer struct {
}

func (t *thumbnailer) Thumbnail(origin io.Reader, width, height, cast int, extension string) (thumbnail io.Reader, later func(io.Reader) (io.Reader, error), err error) {
	originFile, err := ioutil.TempFile("", "thumbnail-or")
	if err != nil {
		return
	}
	defer func(originFile *os.File) {
		_ = originFile.Close()
	}(originFile)
	defer func(name string) {
		_ = os.Remove(name)
	}(originFile.Name())

	_, err = io.Copy(originFile, origin)
	if err != nil {
		return
	}

	destinationFile, err := ioutil.TempFile("", "thumbnail-dst.*."+extension)
	if err != nil {
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(originFile.Name())

	later, err = t.thumbnailFiles(originFile.Name(), destinationFile.Name(), width, height, cast, extension)
	if err != nil {
		return
	}

	thumbnail = bufio.NewReader(destinationFile)
	return
}

func (t *thumbnailer) thumbnailFiles(origin, destination string, width, height, cast int, extension string) (later func(io.Reader) (io.Reader, error), err error) {
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
	saveQuality := quality
	if saveQuality > 100 {
		saveQuality = 100
	}

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
			imagickParams = append(imagickParams, "-layers", "optimize", "-quality", strconv.Itoa(saveQuality))
			err = t.execMagick(origin, destination, imagickParams)
		}
	} else {
		if info.Animated {
			imagickParams = append(imagickParams, "-flatten")
			origin += "[0]"
		}

		if extension == FORMAT_JPG {
			destinationTempFile, er := ioutil.TempFile("", "saveproxy")
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

			_, err = helper.Exec("sh", "-c", "mozjpeg -optimize -progressive -quality "+strconv.Itoa(saveQuality)+" "+shellescape.Quote(destinationTempFile.Name())+" > "+shellescape.Quote(destination))
			if err != nil {
				return
			}

		} else {
			err = t.execMagick(origin, destination, append(imagickParams, "-quality", strconv.Itoa(saveQuality)))
			if err != nil {
				return
			}

			if extension == FORMAT_PNG {
				// minimal compression now
				_, err = helper.Exec("optipng", destination)
				if err != nil {
					return
				}

				later = t.laterOptimizePng
			}
		}
	}

	return
}

func (t *thumbnailer) laterOptimizePng(uncompressed io.Reader) (compressed io.Reader, err error) {
	uncompressedFile, err := ioutil.TempFile("", "thumbnail-pngi")
	defer func(name string) {
		_ = os.Remove(name)
	}(uncompressedFile.Name())

	_, err = io.Copy(uncompressedFile, uncompressed)
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

	_, err = helper.Exec("zopflipng", "--iterations="+strconv.Itoa(quality), "-y", "--filters=01234mepb", "--lossy_8bit", "--lossy_transparent", uncompressedFile.Name(), compressedFile.Name())
	if err != nil {
		return
	}

	compressed = bufio.NewReader(compressedFile)
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

func (t *thumbnailer) getQuality(width, height int, format string) int {
	quality := config.Get().QualityDefault

	halfPerimeter := width + height

	for _, qualityFormat := range config.Get().Quality {
		if qualityFormat.Format == format {
			quality = qualityFormat.Quality
			for _, condition := range qualityFormat.Conditions {
				if halfPerimeter >= condition.From && halfPerimeter < condition.To {
					quality = condition.Quality
					break
				}
			}
			break
		}
	}

	return quality
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
