package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// must be public for toml coding to function
type TomlV2Common struct {
	StoreVersion      store_version.Version `toml:"store-version"`
	_                 string                `toml:"repo-type"`
	RepoId            ids.RepoId            `toml:"id"`
	InventoryListType string                `toml:"inventory_list-type"`
	ObjectSigType     string                `toml:"object-sig-type"`
}

type TomlV2Private struct {
	PrivateKey markl.Id `toml:"private-key"`
	TomlV2Common
}

type TomlV2Public struct {
	PublicKey markl.Id `toml:"public-key"`
	TomlV2Common
}

func (config *TomlV2Common) SetFlagSet(
	flagSet interfaces.CommandLineFlagDefinitions,
) {
}

func (config *TomlV2Common) SetRepoId(id ids.RepoId) {
	config.RepoId = id
}

func (config *TomlV2Common) GetInventoryListTypeId() string {
	if config.InventoryListType == "" {
		return ids.TypeInventoryListV1
	} else {
		return config.InventoryListType
	}
}

func (config *TomlV2Common) GetObjectSigMarklTypeId() string {
	if config.ObjectSigType == "" {
		return markl.PurposeObjectSigV1
	} else {
		return config.ObjectSigType
	}
}

func (config *TomlV2Common) SetInventoryListTypeId(value string) {
	config.InventoryListType = value
}

func (config *TomlV2Common) SetObjectSigMarklTypeId(value string) {
	config.ObjectSigType = value
}

func (config *TomlV2Private) GetGenesisConfig() ConfigPrivate {
	return config
}

func (config *TomlV2Private) GetGenesisConfigPublic() ConfigPublic {
	return &TomlV2Public{
		TomlV2Common: config.TomlV2Common,
		PublicKey:    config.GetPublicKey(),
	}
}

func (config *TomlV2Private) GetPrivateKey() markl.Id {
	return config.PrivateKey
}

func (config *TomlV2Private) GetPrivateKeyMutable() *markl.Id {
	return &config.PrivateKey
}

func (config *TomlV2Private) GetPublicKey() markl.Id {
	public, err := markl.GetPublicKey(config.PrivateKey)
	errors.PanicIfError(err)
	return public
}

func (config *TomlV2Public) GetGenesisConfig() ConfigPublic {
	return config
}

func (config TomlV2Public) GetPublicKey() markl.Id {
	return config.PublicKey
}

func (config *TomlV2Common) GetStoreVersion() store_version.Version {
	return config.StoreVersion
}

func (config TomlV2Common) GetRepoId() ids.RepoId {
	return config.RepoId
}
