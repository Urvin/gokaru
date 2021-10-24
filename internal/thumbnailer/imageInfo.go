package thumbnailer

import (
	"gopkg.in/gographics/imagick.v3/imagick"
)

type ImageInfo struct {
	Width    uint
	Height   uint
	Format   string
	Alpha    bool
	Animated bool
}

func GetImageInfoWand(mw *imagick.MagickWand) (info ImageInfo, err error) {
	info.Width = mw.GetImageWidth()
	info.Height = mw.GetImageHeight()
	info.Format = mw.GetImageFormat()
	info.Alpha = mw.GetImageAlphaChannel()
	info.Animated = mw.GetNumberImages() > 1

	if info.Alpha {
		info.Alpha = false
		tw := int(info.Width)
		th := int(info.Height)

		for x := 0; x < tw; x++ {
			for y := 0; y < th; y++ {
				pw, er := mw.GetImagePixelColor(x, y)
				if er != nil {
					err = er
					return
				}
				if pw.GetAlpha() > 0 {
					info.Alpha = true
					return
				}
			}
		}
	}

	return
}

func GetImageInfoFile(fileName string) (info ImageInfo, err error) {
	mw := imagick.NewMagickWand()
	err = mw.ReadImage(fileName)
	if err != nil {
	}
	info, err = GetImageInfoWand(mw)
	return
}
