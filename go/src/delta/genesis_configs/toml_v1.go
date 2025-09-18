package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// must be public for toml coding to function
type TomlV1Common struct {
	StoreVersion      store_version.Version     `toml:"store-version"`
	_                 string                    `toml:"repo-type"`
	RepoId            ids.RepoId                `toml:"id"`
	BlobStore         blob_store_configs.TomlV0 `toml:"blob-store"`
	InventoryListType string                    `toml:"inventory_list-type"`
}

type TomlV1Private struct {
	PrivateKey markl.IdBroken `toml:"private-key"`
	TomlV1Common
}

var _ ConfigPrivate = &TomlV1Private{}

type TomlV1Public struct {
	PublicKey markl.IdBroken `toml:"public-key"`
	TomlV1Common
}

var _ ConfigPublic = &TomlV1Public{}

func (config *TomlV1Common) SetFlagDefinitions(
	flagSet interfaces.CommandLineFlagDefinitions,
) {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		config.BlobStore.SetFlagDefinitions(flagSet)
	}
}

func (config *TomlV1Common) SetRepoId(id ids.RepoId) {
	config.RepoId = id
}

func (config *TomlV1Common) GetInventoryListTypeId() string {
	if config.InventoryListType == "" {
		return ids.TypeInventoryListV1
	} else {
		return config.InventoryListType
	}
}

func (config *TomlV1Common) SetInventoryListTypeId(value string) {
	config.InventoryListType = value
}

func (config *TomlV1Common) GetObjectSigMarklTypeId() string {
	return markl.PurposeObjectSigV0
}

func (config *TomlV1Common) SetObjectSigMarklTypeId(string) {
	panic(errors.Err405MethodNotAllowed)
}

func (config *TomlV1Private) GetGenesisConfig() ConfigPrivate {
	return config
}

func (config *TomlV1Private) GetGenesisConfigPublic() ConfigPublic {
	public, err := markl.Id(
		config.PrivateKey,
	).GetPublicKey(
		markl.PurposeRepoPrivateKeyV1,
	)
	errors.PanicIfError(err)

	return &TomlV1Public{
		TomlV1Common: config.TomlV1Common,
		PublicKey:    markl.IdBroken(public),
	}
}

func (config *TomlV1Private) GetPrivateKey() interfaces.MarklId {
	return markl.Id(config.PrivateKey)
}

func (config *TomlV1Private) GetPrivateKeyMutable() interfaces.MutableMarklId {
	return (*markl.Id)(&config.PrivateKey)
}

func (config *TomlV1Private) GetPublicKey() interfaces.MarklId {
	public, err := (markl.Id)(
		config.PrivateKey,
	).GetPublicKey(
		markl.PurposeRepoPrivateKeyV1,
	)
	errors.PanicIfError(err)
	return public
}

func (config *TomlV1Public) GetGenesisConfig() ConfigPublic {
	return config
}

func (config TomlV1Public) GetPublicKey() interfaces.MarklId {
	return (markl.Id)(config.PublicKey)
}

func (config *TomlV1Common) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	return &config.BlobStore
}

func (config *TomlV1Common) GetStoreVersion() store_version.Version {
	return config.StoreVersion
}

func (config TomlV1Common) GetRepoId() ids.RepoId {
	return config.RepoId
}
