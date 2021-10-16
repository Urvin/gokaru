package security

type SignatureGenerator interface {
	SetSalt(salt string)
	Sign(sourceType, fileCategory, fileName string, width, height, cast int) string
}
