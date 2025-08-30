package repo_blobs

import (
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/markl_toml"
)

type TomlUriV0 struct {
	markl_toml.TomlPublicKeyV0
	Uri values.Uri `toml:"uri"`
}

func (b TomlUriV0) GetRepoBlob() Blob {
	return b
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
