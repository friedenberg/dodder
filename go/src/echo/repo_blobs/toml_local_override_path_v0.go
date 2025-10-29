package repo_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

type TomlLocalPathV0 struct {
	PublicKey markl.Id `toml:"public-key"`
	Path      string   `toml:"path"`
}

var (
	_ Blob        = TomlLocalPathV0{}
	_ BlobMutable = &TomlLocalPathV0{}
)

func (config TomlLocalPathV0) GetRepoBlob() Blob {
	return config
}

func (config TomlLocalPathV0) GetPublicKey() interfaces.MarklId {
	return config.PublicKey
}

func (config *TomlLocalPathV0) SetPublicKey(id interfaces.MarklId) {
	config.PublicKey.ResetWithMarklId(id)
}

func (config *TomlLocalPathV0) Reset() {
	config.Path = ""
}

func (config *TomlLocalPathV0) ResetWith(b TomlLocalPathV0) {
	config.Path = b.Path
}

func (config TomlLocalPathV0) Equals(b TomlLocalPathV0) bool {
	if config.Path != b.Path {
		return false
	}

	return true
}
