package genesis_config

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type V0Common struct {
	StoreVersion      StoreVersion
	Recipients        []string
	CompressionType   compression_type.CompressionType
	LockInternalFiles bool
}

type V0Public struct {
	V0Common
}

type V0Private struct {
	V0Common
}

func (config *V0Common) SetFlagSet(f *flag.FlagSet) {
	config.CompressionType.SetFlagSet(f)

	f.BoolVar(
		&config.LockInternalFiles,
		"lock-internal-files",
		config.LockInternalFiles,
		"",
	)

	f.Func(
		"recipient",
		"age recipients",
		func(value string) (err error) {
			config.Recipients = append(config.Recipients, value)
			return
		},
	)
}

func (config *V0Public) config() public   { return public{} }
func (config *V0Private) config() private { return private{} }

func (config *V0Private) GetImmutableConfig() Private {
	return config
}

func (config *V0Private) GetImmutableConfigPublic() Public {
	return &V0Public{
		V0Common: config.V0Common,
	}
}

func (config *V0Public) GetImmutableConfigPublic() Public {
	return config
}

func (config *V0Common) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return config
}

func (config V0Common) GetStoreVersion() StoreVersion {
	return config.StoreVersion
}

func (config V0Common) GetRepoType() repo_type.Type {
	return repo_type.TypeWorkingCopy
}

func (config V0Common) GetPrivateKey() repo_signing.PrivateKey {
	panic(errors.ErrorWithStackf("not supported"))
}

func (config V0Common) GetPublicKey() repo_signing.PublicKey {
	panic(errors.ErrorWithStackf("not supported"))
}

func (config V0Common) GetRepoId() ids.RepoId {
	return ids.RepoId{}
}

func (config *V0Common) GetAgeEncryption() *age.Age {
	return &age.Age{}
}

func (config *V0Common) GetBlobCompression() interfaces.BlobCompression {
	return &config.CompressionType
}

func (config *V0Common) GetBlobEncryption() interfaces.BlobEncryption {
	return config.GetAgeEncryption()
}

func (config V0Common) GetLockInternalFiles() bool {
	return config.LockInternalFiles
}

func (config V0Common) GetInventoryListTypeString() string {
	return InventoryListTypeV0
}
