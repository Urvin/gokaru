package thumbnailer

import (
	"github.com/urvin/gokaru/internal/helper"
	"io/ioutil"
	"os"
	"strconv"
)

func pngquant(image []byte, qualityMin uint, qualityMax uint) (result []byte, err error) {
	tfi, err := ioutil.TempFile("", "pngquant_in_*.png")
	if err != nil {
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tfi.Name())

	tfo, err := ioutil.TempFile("", "pngquant_out_*.out")
	if err != nil {
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tfo.Name())

	err = ioutil.WriteFile(tfi.Name(), image, 0644)
	if err != nil {
		return
	}

	_, err = helper.Exec("pngquant", "--force",
		"--quality="+strconv.FormatUint(uint64(qualityMin), 10)+"-"+strconv.FormatUint(uint64(qualityMax), 10),
		"--strip", "--skip-if-larger", "--speed", "1",
		"--output", ".png",
		"--verbose", "--", tfo.Name())
	if err != nil {
		return
	}

	result, err = ioutil.ReadFile(tfo.Name())

	return
}
