package thumbnailer

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"gopkg.in/alessio/shellescape.v1"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

func Thumbnail(origin, destination string, width, height, cast int, extension string) (err error) {
	err = validate(width, height, cast, extension)
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
		hasCast(CAST_TRANSPARENT_BACKGROUND, cast) &&
		!hasCast(CAST_TRANSPARENT_BACKGROUND, cast) &&
		!destinationIsOpaque {
		firstPixelColor, er := getFirstPixelColor(origin)
		if er != nil {
			err = er
			return
		}
		if firstPixelColor != "" {
			imagickParams = append(imagickParams, "-alpha", "off", "-bordercolor", firstPixelColor, "-border", "1", "(", "+clone", "-fuzz", "20%", "-fill", "none", "-floodfill", "+0+0", firstPixelColor, "-alpha", "extract", "-geometry", "200%", "-blur", "0x0.5", "-morphology", "erode", "square:1", "-geometry", "50%", ")", "-compose", "CopyOpacity", "-composite", "-shave", "1", "-compose", "add")
		}
	}

	// Set opaque background
	if hasCast(CAST_OPAQUE_BACKGROUND, cast) || (info.Alpha && destinationIsOpaque) {
		imagickParams = append(imagickParams, "-background", "white", "-alpha", "remove", "-alpha", "off", "-fuzz", "5%", "-fill", "white", "-draw", "color 1,1 floodfill")
		opaqueBackgroundSet = true
	}

	// trim image
	if hasCast(CAST_TRIM, cast) {
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
	} else if hasCast(CAST_RESIZE_TENSILE, cast) {
		resizeFlag = "!"
	} else if hasCast(CAST_RESIZE_PRECISE, cast) {
		resizeFlag = "^"
	} else if hasCast(CAST_RESIZE_INVERSE, cast) {
		resizeFlag = ""
	} else {
		forceExtent = true
	}

	resizeWidth := width
	resizeHeight := height
	padding := config.Get().Padding
	if hasCast(CAST_TRIM_PADDING, cast) &&
		hasCast(CAST_TRIM, cast) &&
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
	if forceExtent || hasCast(CAST_EXTENT, cast) {
		if !opaqueBackgroundSet {
			imagickParams = append(imagickParams, "-background", "none")
		}
		imagickParams = append(imagickParams, "-gravity", "center", "-extent", strconv.Itoa(width)+"x"+strconv.Itoa(height))
	}

	quality := getQuality(width, height, extension)

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
			defer os.Remove(paletteFile.Name())

			err = execMagick(origin, FORMAT_GIF+":"+paletteFile.Name(), paletteParams)
			if err != nil {
				return
			}

			imagickParams = append(imagickParams, "+dither", "-remap", paletteFile.Name(), "-layers", "optimize")
			err = execMagick(origin, destination, imagickParams)
			if err != nil {
				return
			}
		} else {
			imagickParams = append(imagickParams, "-layers", "optimize", "-quality", strconv.Itoa(quality))
			err = execMagick(origin, destination, imagickParams)
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
			defer os.Remove(destinationTempFile.Name())

			err = execMagick(origin, FORMAT_PNG+":"+destinationTempFile.Name(), append(imagickParams, "-quality", "100"))
			if err != nil {
				return err
			}

			//_, err = helper.Exec("mozjpeg", "-optimize", "-progressive", "-quality", strconv.Itoa(quality), destinationTempFile.Name(), ">", destination)
			_, err = helper.Exec("sh", "-c", "mozjpeg -optimize -progressive -quality "+strconv.Itoa(quality)+" "+shellescape.Quote(destinationTempFile.Name())+" > "+shellescape.Quote(destination))
			if err != nil {
				return
			}

		} else {
			saveQuality := quality
			if saveQuality > 100 {
				saveQuality = 100
			}
			err = execMagick(origin, destination, append(imagickParams, "-quality", strconv.Itoa(saveQuality)))
			if err != nil {
				return err
			}

			if extension == FORMAT_PNG {
				// minimal compression now
				_, err = helper.Exec("optipng", destination)
				if err != nil {
					return
				}

				// maximum later
				go func() {
					log.Info("Thubmnailer: zopfli started")
					_, er := helper.Exec("zopflipng", "-m", "--iterations="+strconv.Itoa(quality), "-y", "--lossy_transparent", destination, destination)
					if er != nil {
						log.Warn("Thubmnailer: zopfli error:" + er.Error())
					}
					log.Info("Thubmnailer: zopfli ended")
				}()
			}
		}
	}
	return
}

func validate(width, height, cast int, extension string) (err error) {
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

func hasCast(needle int, haystack int) bool {
	return (haystack & needle) > 0
}

func getFirstPixelColor(fileName string) (color string, err error) {

	output, er := helper.Exec("convert", fileName, "-format", "%[pixel:p{0,0}]", "info:-")
	if er != nil {
		err = er
		return
	}
	color = strings.Trim(output, "\n\r\t ")
	return
}

func getQuality(width, height int, format string) int {
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

	log.Info("Thubmnailer: halfperimeter " + strconv.Itoa(halfPerimeter) + " format " + format + " quality " + strconv.Itoa(quality))

	return quality
}

func execMagick(origin, destination string, params []string) (err error) {
	params = append([]string{origin}, append(params, destination)...)
	_, err = helper.Exec("convert", params...)
	if err != nil {
		return
	}
	return
}
