package blob_store_configs

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

// TODO add path option
type TomlV0 struct {
	HashBuckets       []int                            `toml:"hash-buckets"`
	AgeEncryption     age.Age                          `toml:"age-encryption,omitempty"`
	CompressionType   compression_type.CompressionType `toml:"compression-type"`
	LockInternalFiles bool                             `toml:"lock-internal-files"`
}

func (*TomlV0) GetBlobStoreType() string {
	return "local"
}

func (blobStoreConfig *TomlV0) SetFlagSet(flagSet *flag.FlagSet) {
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
	return blobStoreConfig.HashBuckets
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
