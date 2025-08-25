package merkle

import (
	"crypto"

	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

// TODO hide inner fields
type TomlPublicKeyV0 struct {
	PublicKey blech32.Value `toml:"public-key,omitempty"`
}

func (b TomlPublicKeyV0) GetPublicKey() PublicKey {
	return b.PublicKey.Data
}

func (b *TomlPublicKeyV0) SetPublicKey(key crypto.PublicKey) {
	b.PublicKey.HRP = HRPRepoPubKeyV1
	b.PublicKey.Data = key.(PublicKey)
}
