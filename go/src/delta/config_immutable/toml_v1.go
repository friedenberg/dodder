package config_immutable

import (
	"crypto/ed25519"
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type TomlV1Common struct {
	StoreVersion StoreVersion    `toml:"store-version"`
	RepoType     repo_type.Type  `toml:"repo-type"`
	RepoId       ids.RepoId      `toml:"id"`
	BlobStore    BlobStoreTomlV1 `toml:"blob-store"`
}

type TomlV1Private struct {
	repo_signing.TomlPrivateKeyV0
	TomlV1Common
}

type TomlV1Public struct {
	repo_signing.TomlPublicKeyV0
	TomlV1Common
}

func (k *TomlV1Common) SetFlagSet(f *flag.FlagSet) {
	k.BlobStore.SetFlagSet(f)
	k.RepoType = repo_type.TypeWorkingCopy
	f.Var(&k.RepoType, "repo-type", "")
}

func (k *TomlV1Public) config() public   { return public{} }
func (k *TomlV1Private) config() private { return private{} }

func (k *TomlV1Private) GetImmutableConfig() ConfigPrivate {
	return k
}

func (k *TomlV1Private) GetImmutableConfigPublic() ConfigPublic {
	return &TomlV1Public{
		TomlV1Common:    k.TomlV1Common,
		TomlPublicKeyV0: k.TomlPrivateKeyV0.GetPublicKey(),
	}
}

func (k *TomlV1Private) GetPrivateKey() repo_signing.PrivateKey {
	return repo_signing.NewKeyFromSeed(k.PrivateKey.Data)
}

func (k *TomlV1Private) GetPublicKey() repo_signing.PublicKey {
	return repo_signing.PublicKey(k.GetPrivateKey().Public().(ed25519.PublicKey))
}

func (k *TomlV1Public) GetImmutableConfigPublic() ConfigPublic {
	return k
}

func (k TomlV1Public) GetPublicKey() repo_signing.PublicKey {
	return k.PublicKey.Data
}

func (k *TomlV1Common) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return &k.BlobStore
}

func (k *TomlV1Common) GetStoreVersion() interfaces.StoreVersion {
	return k.StoreVersion
}

func (k TomlV1Common) GetRepoType() repo_type.Type {
	return k.RepoType
}

func (k TomlV1Common) GetRepoId() ids.RepoId {
	return k.RepoId
}
