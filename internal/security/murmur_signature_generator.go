package security

import (
	"github.com/spaolacci/murmur3"
	"github.com/urvin/gokaru/internal/contracts"
	"strconv"
)

type murmurSignatureGenerator struct {
	salt string
}

func (sg *murmurSignatureGenerator) SetSalt(salt string) {
	sg.salt = salt
}

func (sg *murmurSignatureGenerator) Sign(miniature *contracts.MiniatureDto) string {
	hash := murmur3.New32()
	_, err := hash.Write([]byte(sg.salt + "/" + miniature.Hash()))
	if err != nil {
		return ""
	}
	return strconv.FormatUint(uint64(hash.Sum32()), 32)
}

func NewMurmurSignatureGenerator() SignatureGenerator {
	result := &murmurSignatureGenerator{}
	return result
}
