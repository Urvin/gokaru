package thumbnailer

import (
	"github.com/urvin/gokaru/internal/helper"
	"github.com/urvin/gokaru/internal/vips"
	"io/ioutil"
	"os"
	"strings"
)

// My fault, really don't know how to run it with vips or magick lib
func createTransparentBackground(image *vips.Image) (err error) {
	in, err := ioutil.TempFile("", "transbg-in-*.png")
	if err != nil {
		return err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(in.Name())

	out, err := ioutil.TempFile("", "transbg-out-*.png")
	if err != nil {
		return err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(out.Name())

	po := vips.NewPngSaveOptions()
	po.Compression = 1
	po.Quantize = false
	po.Interlace = false
	ind, err := image.SavePng(po)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(in.Name(), ind, 0644)
	if err != nil {
		return err
	}

	fpc, err := getFirstPixelColor(in.Name())
	if err != nil {
		return err
	}

	imagickParams := []string{"-alpha", "off", "-bordercolor", fpc, "-border", "1", "(", "+clone", "-fuzz", "20%", "-fill", "none", "-floodfill", "+0+0", fpc, "-alpha", "extract", "-geometry", "200%", "-blur", "0x0.5", "-morphology", "erode", "square:1", "-geometry", "50%", ")", "-compose", "CopyOpacity", "-composite", "-shave", "1", "-compose", "add"}
	err = execMagick(in.Name(), out.Name(), imagickParams)
	if err != nil {
		return err
	}

	oud, err := ioutil.ReadFile(out.Name())
	if err != nil {
		return
	}

	image.Clear()

	err = image.Load(oud, vips.ImageTypePNG, 1, 1.0, 1)
	if err != nil {
		return
	}

	return
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

func execMagick(origin, destination string, params []string) (err error) {
	params = append([]string{origin}, append(params, destination)...)
	_, err = helper.Exec("convert", params...)
	if err != nil {
		return
	}
	return
}
