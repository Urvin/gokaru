package helper

import (
	"strconv"
	"strings"
)

func FileNameWithoutExtension(fileName string) string {
	if pos := strings.LastIndexByte(fileName, '.'); pos != -1 {
		return fileName[:pos]
	}
	return fileName
}

func Atoi(input string) int {
	i, err := strconv.Atoi(input)
	if err != nil {
		return 0
	}
	return i
}
