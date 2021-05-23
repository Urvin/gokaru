package security

import (
	"github.com/urvin/gokaru/internal/helper"
	"strconv"
)

func signatureSalt() string {
	return "nocheck"
}

func GenerateSignature(sourceType, fileCategory, fileName string, width, height, cast int) string {
	return helper.MD5(signatureSalt() + "/" + sourceType + "/" + fileCategory + "/" + fileName + "/" +
		strconv.Itoa(width) + "/" + strconv.Itoa(height) + "/" + strconv.Itoa(cast))
}
