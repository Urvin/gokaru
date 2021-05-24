package security

import (
	"github.com/urvin/gokaru/internal/config"
	"github.com/urvin/gokaru/internal/helper"
	"strconv"
)

func signatureSalt() string {
	salt := config.Get().SignatureSalt
	return salt
}

func GenerateSignature(sourceType, fileCategory, fileName string, width, height, cast int) string {
	return helper.MD5(signatureSalt() + "/" + sourceType + "/" + fileCategory + "/" + fileName + "/" +
		strconv.Itoa(width) + "/" + strconv.Itoa(height) + "/" + strconv.Itoa(cast))
}
