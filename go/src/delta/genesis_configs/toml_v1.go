package genesis_configs

import (
	"crypto/ed25519"
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// must be public for toml coding to function
type TomlV1Common struct {
	StoreVersion      store_version.Version     `toml:"store-version"`
	RepoType          repo_type.Type            `toml:"repo-type"`
	RepoId            ids.RepoId                `toml:"id"`
	BlobStore         blob_store_configs.TomlV0 `toml:"blob-store"`
	InventoryListType string                    `toml:"inventory_list-type"`
}

type TomlV1Private struct {
	repo_signing.TomlPrivateKeyV0
	TomlV1Common
}

type TomlV1Public struct {
	repo_signing.TomlPublicKeyV0
	TomlV1Common
}

func (config *TomlV1Common) SetFlagSet(flagSet *flag.FlagSet) {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		config.BlobStore.SetFlagSet(flagSet)
	}
	config.RepoType = repo_type.TypeWorkingCopy
	flagSet.Var(&config.RepoType, "repo-type", "")
}

func (config *TomlV1Common) SetRepoType(tipe repo_type.Type) {
	config.RepoType = tipe
}

func (config *TomlV1Common) SetRepoId(id ids.RepoId) {
	config.RepoId = id
}

func (config *TomlV1Common) GetInventoryListTypeString() string {
	if config.InventoryListType == "" {
		return ids.TypeInventoryListV1
	} else {
		return config.InventoryListType
	}
}

func (config *TomlV1Common) SetInventoryListTypeString(value string) {
	config.InventoryListType = value
}

func (config *TomlV1Private) GetGenesisConfig() ConfigPrivate {
	return config
}

func (config *TomlV1Private) GetGenesisConfigPublic() ConfigPublic {
	return &TomlV1Public{
		TomlV1Common:    config.TomlV1Common,
		TomlPublicKeyV0: config.TomlPrivateKeyV0.GetPublicKey(),
	}
}

func (config *TomlV1Private) GetPrivateKey() repo_signing.PrivateKey {
	return repo_signing.NewKeyFromSeed(config.PrivateKey.Data)
}

func (config *TomlV1Private) GetPublicKey() repo_signing.PublicKey {
	return repo_signing.PublicKey(
		config.GetPrivateKey().Public().(ed25519.PublicKey),
	)
}

func (config *TomlV1Public) GetGenesisConfig() ConfigPublic {
	return config
}

func (config TomlV1Public) GetPublicKey() repo_signing.PublicKey {
	return config.PublicKey.Data
}

func (config *TomlV1Common) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	return &config.BlobStore
}

func (config *TomlV1Common) GetStoreVersion() store_version.Version {
	return config.StoreVersion
}

func (config TomlV1Common) GetRepoType() repo_type.Type {
	return config.RepoType
}

func (config TomlV1Common) GetRepoId() ids.RepoId {
	return config.RepoId
}
