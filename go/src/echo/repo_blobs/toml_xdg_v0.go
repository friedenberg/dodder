package repo_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type TomlXDGV0 struct {
	PublicKey markl.Id `toml:"public-key"`
	Data      string   `toml:"data"`
	Config    string   `toml:"config"`
	State     string   `toml:"state"`
	Cache     string   `toml:"cache"`
	Runtime   string   `toml:"runtime"`
}

var (
	_ Blob        = TomlXDGV0{}
	_ BlobMutable = &TomlXDGV0{}
)

func TomlXDGV0FromXDG(xdg xdg.XDG) *TomlXDGV0 {
	return &TomlXDGV0{
		Data:    xdg.Data.String(),
		Config:  xdg.Config.String(),
		State:   xdg.State.String(),
		Cache:   xdg.Cache.String(),
		Runtime: xdg.Runtime.String(),
	}
}

func (blob TomlXDGV0) MakeXDG() xdg.XDG {
	return xdg.XDG{
		Data:    xdg.DefaultData.MakeBaseEnvVar(blob.Data),
		Config:  xdg.DefaultConfig.MakeBaseEnvVar(blob.Config),
		Cache:   xdg.DefaultCache.MakeBaseEnvVar(blob.Cache),
		Runtime: xdg.DefaultRuntime.MakeBaseEnvVar(blob.Runtime),
		State:   xdg.DefaultState.MakeBaseEnvVar(blob.State),
	}
}

func (blob TomlXDGV0) GetRepoBlob() Blob {
	return blob
}

func (blob TomlXDGV0) GetPublicKey() interfaces.MarklId {
	return blob.PublicKey
}

func (blob *TomlXDGV0) SetPublicKey(id interfaces.MarklId) {
	blob.PublicKey.ResetWithMarklId(id)
}
