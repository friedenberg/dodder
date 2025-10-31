package repo_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

type TomlUriV0 struct {
	PublicKey markl.Id   `toml:"public-key"`
	Uri       values.Uri `toml:"uri"`
}

var (
	_ BlobUri     = TomlUriV0{}
	_ BlobMutable = &TomlUriV0{}
)

func (b TomlUriV0) GetRepoBlob() Blob {
	return b
}

func (config TomlUriV0) GetPublicKey() interfaces.MarklId {
	return config.PublicKey
}

func (config *TomlUriV0) SetPublicKey(id interfaces.MarklId) {
	config.PublicKey.ResetWithMarklId(id)
}

func (b TomlUriV0) GetRepoType() {
}

func (a *TomlUriV0) Reset() {
	a.Uri = values.Uri{}
}

func (a *TomlUriV0) ResetWith(b TomlUriV0) {
	a.Uri = b.Uri
}

func (a TomlUriV0) Equals(b TomlUriV0) bool {
	if a.Uri != b.Uri {
		return false
	}

	return true
}

func (config TomlUriV0) IsRemote() bool {
	return true
}

func (config TomlUriV0) GetUri() values.Uri {
	return config.Uri
}
