package repo_signing

import (
	"crypto"

	"code.linenisgreat.com/dodder/go/src/bravo/bech32"
)

// TODO hide inner fields
type TomlPublicKeyV0 struct {
	PublicKey bech32.Value `toml:"public-key,omitempty"`
}

func (b TomlPublicKeyV0) GetPublicKey() PublicKey {
	return b.PublicKey.Data
}

func (b *TomlPublicKeyV0) SetPublicKey(key crypto.PublicKey) {
	b.PublicKey.HRP = HRPRepoPubKeyV1
	b.PublicKey.Data = key.(PublicKey)
}
