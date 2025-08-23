package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type TomlV0 struct {
	// TODO add hasher option
	// TODO add path option
	// TODO uncomment when bumping to V1
	// HashBuckets       []int                            `toml:"hash-buckets"`
	AgeEncryption     age.Age                          `toml:"age-encryption,omitempty"`
	CompressionType   compression_type.CompressionType `toml:"compression-type"`
	LockInternalFiles bool                             `toml:"lock-internal-files"`
}

func (*TomlV0) GetBlobStoreType() string {
	return "local"
}

func (blobStoreConfig *TomlV0) SetFlagSet(flagSet *flags.FlagSet) {
	blobStoreConfig.CompressionType.SetFlagSet(flagSet)

	flagSet.BoolVar(
		&blobStoreConfig.LockInternalFiles,
		"lock-internal-files",
		blobStoreConfig.LockInternalFiles,
		"",
	)

	flagSet.Var(
		&blobStoreConfig.AgeEncryption,
		"age-identity",
		"add an age identity",
	)
}

func (blobStoreConfig *TomlV0) GetHashBuckets() []int {
	return []int{2}
}

func (blobStoreConfig *TomlV0) GetBlobCompression() interfaces.BlobCompression {
	return &blobStoreConfig.CompressionType
}

func (blobStoreConfig *TomlV0) GetBlobEncryption() interfaces.BlobEncryption {
	return &blobStoreConfig.AgeEncryption
}

func (blobStoreConfig *TomlV0) GetLockInternalFiles() bool {
	return blobStoreConfig.LockInternalFiles
}
