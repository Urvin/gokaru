package security

import (
	"github.com/spaolacci/murmur3"
	"strconv"
)

type murmurSignatureGenerator struct {
	salt string
}

func (sg *murmurSignatureGenerator) SetSalt(salt string) {
	sg.salt = salt
}

func (sg *murmurSignatureGenerator) Sign(sourceType, fileCategory, fileName string, width, height, cast int) string {
	hash := murmur3.New32()
	_, err := hash.Write([]byte(
		sg.salt + "/" +
			sourceType + "/" +
			fileCategory + "/" +
			fileName + "/" +
			strconv.Itoa(width) + "/" +
			strconv.Itoa(height) + "/" +
			strconv.Itoa(cast),
	))
	if err != nil {
		return ""
	}
	return strconv.FormatUint(uint64(hash.Sum32()), 32)
}

func NewMurmurSignatureGenerator() SignatureGenerator {
	result := &murmurSignatureGenerator{}
	return result
}
