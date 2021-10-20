package security

import "github.com/urvin/gokaru/internal/contracts"

type SignatureGenerator interface {
	SetSalt(salt string)
	Sign(miniature *contracts.Miniature) string
}
