package genesis_config

import (
	"crypto/ed25519"
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// must be public for toml coding to function
type TomlV2Common struct {
	StoreVersion      StoreVersion   `toml:"store-version"`
	RepoType          repo_type.Type `toml:"repo-type"`
	RepoId            ids.RepoId     `toml:"id"`
	InventoryListType string         `toml:"inventory_list-type"`
}

type TomlV2Private struct {
	repo_signing.TomlPrivateKeyV0
	TomlV2Common
}

type TomlV2Public struct {
	repo_signing.TomlPublicKeyV0
	TomlV2Common
}

func (config *TomlV2Common) SetFlagSet(flagSet *flag.FlagSet) {
	config.RepoType = repo_type.TypeWorkingCopy
	flagSet.Var(&config.RepoType, "repo-type", "")
}

func (config *TomlV2Common) SetRepoType(tipe repo_type.Type) {
	config.RepoType = tipe
}

func (config *TomlV2Common) SetRepoId(id ids.RepoId) {
	config.RepoId = id
}

func (config *TomlV2Common) GetInventoryListTypeString() string {
	if config.InventoryListType == "" {
		return InventoryListTypeV1
	} else {
		return config.InventoryListType
	}
}

func (config *TomlV2Public) config() public   { return public{} }
func (config *TomlV2Private) config() private { return private{} }

func (config *TomlV2Private) GetImmutableConfig() Private {
	return config
}

func (config *TomlV2Private) GetImmutableConfigPublic() Public {
	return &TomlV2Public{
		TomlV2Common:    config.TomlV2Common,
		TomlPublicKeyV0: config.TomlPrivateKeyV0.GetPublicKey(),
	}
}

func (config *TomlV2Private) GetPrivateKey() repo_signing.PrivateKey {
	return repo_signing.NewKeyFromSeed(config.PrivateKey.Data)
}

func (config *TomlV2Private) GetPublicKey() repo_signing.PublicKey {
	return repo_signing.PublicKey(
		config.GetPrivateKey().Public().(ed25519.PublicKey),
	)
}

func (config *TomlV2Public) GetImmutableConfigPublic() Public {
	return config
}

func (config TomlV2Public) GetPublicKey() repo_signing.PublicKey {
	return config.PublicKey.Data
}

func (config *TomlV2Common) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return nil
}

func (config *TomlV2Common) GetStoreVersion() StoreVersion {
	return config.StoreVersion
}

func (config TomlV2Common) GetRepoType() repo_type.Type {
	return config.RepoType
}

func (config TomlV2Common) GetRepoId() ids.RepoId {
	return config.RepoId
}
