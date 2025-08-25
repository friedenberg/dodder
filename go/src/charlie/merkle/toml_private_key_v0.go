package merkle

import (
	"crypto"
	"crypto/ed25519"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

// TODO hide inner fields
type TomlPrivateKeyV0 struct {
	PrivateKey blech32.Value `toml:"private-key,omitempty"`
}

func (b *TomlPrivateKeyV0) GeneratePrivateKey() (err error) {
	if len(b.PrivateKey.Data) > 0 {
		err = errors.ErrorWithStackf("private key data already exists, refusing to generate.")
		return
	}

	var privateKey PrivateKey

	if err = privateKey.Generate(nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	b.PrivateKey.Data = privateKey.Seed()
	b.PrivateKey.HRP = HRPRepoPrivateKeyV1

	return
}

func (b TomlPrivateKeyV0) GetPrivateKey() PrivateKey {
	return NewKeyFromSeed(b.PrivateKey.Data)
}

func (b *TomlPrivateKeyV0) SetPrivateKey(key crypto.PrivateKey) {
	b.PrivateKey.HRP = HRPRepoPrivateKeyV1
	b.PrivateKey.Data = key.(PrivateKey)
}

func (b *TomlPrivateKeyV0) GetPublicKey() TomlPublicKeyV0 {
	pub := blech32.Value{
		HRP:  HRPRepoPubKeyV1,
		Data: b.GetPrivateKey().Public().(ed25519.PublicKey),
	}

	return TomlPublicKeyV0{PublicKey: pub}
}
