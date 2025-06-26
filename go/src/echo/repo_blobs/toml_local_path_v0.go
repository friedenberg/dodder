package repo_blobs

import "code.linenisgreat.com/dodder/go/src/charlie/repo_signing"

type TomlLocalPathV0 struct {
	repo_signing.TomlPublicKeyV0
	Path string `toml:"path"`
}

func (b TomlLocalPathV0) GetRepoBlob() Blob {
	return b
}

func (a *TomlLocalPathV0) Reset() {
	a.Path = ""
}

func (a *TomlLocalPathV0) ResetWith(b TomlLocalPathV0) {
	a.Path = b.Path
}

func (a TomlLocalPathV0) Equals(b TomlLocalPathV0) bool {
	if a.Path != b.Path {
		return false
	}

	return true
}
