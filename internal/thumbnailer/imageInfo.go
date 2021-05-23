package thumbnailer

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ImageInfo struct {
	Width    int
	Height   int
	Format   string
	Alpha    bool
	Animated bool
}

func GetImageInfo(fileName string) (info ImageInfo, err error) {
	info.Width = 0
	info.Height = 0
	info.Format = ""
	info.Alpha = false
	info.Animated = false

	cmd := exec.Command("identify", "-format", "%m|%w|%h|%[opaque]++", fileName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	outputString := strings.Trim(string(output), "\n\r\t ")
	imagesData := strings.SplitN(outputString, "++", 2)

	if len(imagesData) > 1 && imagesData[1] != "" {
		info.Animated = true
	}

	data := strings.SplitN(imagesData[0], "|", 4)

	info.Format = strings.ToLower(data[0])
	info.Width, err = strconv.Atoi(data[1])
	if err != nil {
		return
	}
	info.Height, err = strconv.Atoi(data[2])
	if err != nil {
		return
	}
	info.Alpha = data[3] == "true"

	return
}

func getAnimatedWebpImageInfo(fileName string) (info ImageInfo, err error) {
	info.Width = 0
	info.Height = 0
	info.Format = ""
	info.Alpha = false
	info.Animated = true

	temporaryFile, err := ioutil.TempFile("", "webpmux")
	if err != nil {
		return
	}
	cmd := exec.Command("webpmux", "-get", "frame", "1", fileName, "-o", temporaryFile.Name())
	err = cmd.Run()
	if err != nil {
		return
	}

	info, err = GetImageInfo(temporaryFile.Name())
	if err != nil {
		return
	}

	defer os.Remove(temporaryFile.Name())

	info.Animated = true
	info.Format = "webp"
	return
}
