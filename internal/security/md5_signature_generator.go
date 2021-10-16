package security

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

type md5SignatureGenerator struct {
	salt string
}

func (sg *md5SignatureGenerator) SetSalt(salt string) {
	sg.salt = salt
}

func (sg *md5SignatureGenerator) Sign(sourceType, fileCategory, fileName string, width, height, cast int) string {
	hash := md5.Sum([]byte(
		sg.salt + "/" +
			sourceType + "/" +
			fileCategory + "/" +
			fileName + "/" +
			strconv.Itoa(width) + "/" +
			strconv.Itoa(height) + "/" +
			strconv.Itoa(cast),
	))
	return hex.EncodeToString(hash[:])
}

func NewMd5SignatureGenerator() SignatureGenerator {
	result := &md5SignatureGenerator{}
	return result
}
