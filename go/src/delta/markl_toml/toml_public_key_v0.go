package markl_toml

import (
	"crypto"

	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

// TODO hide inner fields
type TomlPublicKeyV0 struct {
	PublicKey blech32.Value `toml:"public-key,omitempty"`
}

func (b TomlPublicKeyV0) GetPublicKey() PublicKey {
	return b.PublicKey.Data
}

func (b *TomlPublicKeyV0) SetPublicKey(key crypto.PublicKey) {
	b.PublicKey.HRP = markl.FormatIdRepoPubKeyV1
	b.PublicKey.Data = key.(PublicKey)
}
