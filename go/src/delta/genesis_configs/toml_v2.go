package genesis_configs

import (
	"crypto/ed25519"

	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// must be public for toml coding to function
type TomlV2Common struct {
	StoreVersion      store_version.Version `toml:"store-version"`
	RepoType          repo_type.Type        `toml:"repo-type"`
	RepoId            ids.RepoId            `toml:"id"`
	InventoryListType string                `toml:"inventory_list-type"`
	ObjectSigType     string                `toml:"object-sig-type"`
}

type TomlV2Private struct {
	merkle.TomlPrivateKeyV0
	TomlV2Common
}

type TomlV2Public struct {
	merkle.TomlPublicKeyV0
	TomlV2Common
}

func (config *TomlV2Common) SetFlagSet(flagSet *flags.FlagSet) {
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
		return ids.TypeInventoryListV1
	} else {
		return config.InventoryListType
	}
}

func (config *TomlV2Common) GetObjectSigTypeString() string {
	if config.ObjectSigType == "" {
		return merkle.HRPObjectSigV1
	} else {
		return config.ObjectSigType
	}
}

func (config *TomlV2Common) SetInventoryListTypeString(value string) {
	config.InventoryListType = value
}

func (config *TomlV2Common) SetObjectSigTypeString(value string) {
	config.ObjectSigType = value
}

func (config *TomlV2Private) GetGenesisConfig() ConfigPrivate {
	return config
}

func (config *TomlV2Private) GetGenesisConfigPublic() ConfigPublic {
	return &TomlV2Public{
		TomlV2Common:    config.TomlV2Common,
		TomlPublicKeyV0: config.TomlPrivateKeyV0.GetPublicKey(),
	}
}

func (config *TomlV2Private) GetPrivateKey() merkle.PrivateKey {
	return merkle.NewKeyFromSeed(config.PrivateKey.Data)
}

func (config *TomlV2Private) GetPublicKey() merkle.PublicKey {
	return merkle.PublicKey(
		config.GetPrivateKey().Public().(ed25519.PublicKey),
	)
}

func (config *TomlV2Public) GetGenesisConfig() ConfigPublic {
	return config
}

func (config TomlV2Public) GetPublicKey() merkle.PublicKey {
	return config.PublicKey.Data
}

func (config *TomlV2Common) GetStoreVersion() store_version.Version {
	return config.StoreVersion
}

func (config TomlV2Common) GetRepoType() repo_type.Type {
	return config.RepoType
}

func (config TomlV2Common) GetRepoId() ids.RepoId {
	return config.RepoId
}
