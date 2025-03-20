package security

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/urvin/gokaru/internal/contracts"
)

type md5SignatureGenerator struct {
	salt string
}

func (sg *md5SignatureGenerator) SetSalt(salt string) {
	sg.salt = salt
}

func (sg *md5SignatureGenerator) Sign(miniature *contracts.MiniatureDto) string {
	hash := md5.Sum([]byte(sg.salt + "/" + miniature.Hash()))
	return hex.EncodeToString(hash[:])
}

func NewMd5SignatureGenerator() SignatureGenerator {
	result := &md5SignatureGenerator{}
	return result
}
