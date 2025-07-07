package config_immutable

import (
	"crypto/ed25519"
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// TODO should this be private?
type TomlV1Common struct {
	StoreVersion      StoreVersion    `toml:"store-version"`
	RepoType          repo_type.Type  `toml:"repo-type"`
	RepoId            ids.RepoId      `toml:"id"`
	BlobStore         BlobStoreTomlV1 `toml:"blob-store"`
	InventoryListType string          `toml:"inventory_list-type"`
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
	config.BlobStore.SetFlagSet(flagSet)
	config.RepoType = repo_type.TypeWorkingCopy
	flagSet.Var(&config.RepoType, "repo-type", "")
}

func (config *TomlV1Common) GetInventoryListTypeString() string {
	if config.InventoryListType == "" {
		return InventoryListTypeV1
	} else {
		return config.InventoryListType
	}
}

func (config *TomlV1Public) config() public   { return public{} }
func (config *TomlV1Private) config() private { return private{} }

func (config *TomlV1Private) GetImmutableConfig() Private {
	return config
}

func (config *TomlV1Private) GetImmutableConfigPublic() Public {
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

func (config *TomlV1Public) GetImmutableConfigPublic() Public {
	return config
}

func (config TomlV1Public) GetPublicKey() repo_signing.PublicKey {
	return config.PublicKey.Data
}

func (config *TomlV1Common) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return &config.BlobStore
}

func (config *TomlV1Common) GetStoreVersion() interfaces.StoreVersion {
	return config.StoreVersion
}

func (config TomlV1Common) GetRepoType() repo_type.Type {
	return config.RepoType
}

func (config TomlV1Common) GetRepoId() ids.RepoId {
	return config.RepoId
}
