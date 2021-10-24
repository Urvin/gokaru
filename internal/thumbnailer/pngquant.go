package thumbnailer

import (
	"github.com/ultimate-guitar/go-imagequant"
	"image/png"
)

func pngquant(image []byte, qualityMin uint, qualityMax uint) (result []byte, err error) {
	result, err = imagequant.Crush(image, 1, png.BestCompression)
	return
}
