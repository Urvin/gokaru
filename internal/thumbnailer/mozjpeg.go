package thumbnailer

import (
	"github.com/urvin/gokaru/internal/helper"
	"gopkg.in/alessio/shellescape.v1"
	"io/ioutil"
	"os"
	"strconv"
)

func mozjpeg(image []byte, quality uint) (result []byte, err error) {
	tfi, err := ioutil.TempFile("", "mozjpeg-in-*.jpg")
	if err != nil {
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tfi.Name())

	tfo, err := ioutil.TempFile("", "mozjpeg-out-*.jpg")
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

	_, err = helper.Exec("sh", "-c", "mozjpeg -optimize -progressive -quality "+strconv.FormatUint(uint64(quality), 10)+" "+shellescape.Quote(tfi.Name())+" > "+shellescape.Quote(tfo.Name()))
	if err != nil {
		return
	}

	result, err = ioutil.ReadFile(tfo.Name())

	return
}
