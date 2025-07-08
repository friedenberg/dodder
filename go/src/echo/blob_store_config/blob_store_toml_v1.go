package blob_store_config

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

// TODO move this to the new package blob_store_config
type BlobStoreTomlV1 struct {
	AgeEncryption     age.Age                          `toml:"age-encryption,omitempty"`
	CompressionType   compression_type.CompressionType `toml:"compression-type"`
	LockInternalFiles bool                             `toml:"lock-internal-files"`
}

func (blobStoreConfig *BlobStoreTomlV1) SetFlagSet(flagSet *flag.FlagSet) {
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

func (blobStoreConfig *BlobStoreTomlV1) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return blobStoreConfig
}

func (blobStoreConfig *BlobStoreTomlV1) GetBlobCompression() interfaces.BlobCompression {
	return &blobStoreConfig.CompressionType
}

func (blobStoreConfig *BlobStoreTomlV1) GetBlobEncryption() interfaces.BlobEncryption {
	return &blobStoreConfig.AgeEncryption
}

func (blobStoreConfig *BlobStoreTomlV1) GetLockInternalFiles() bool {
	return blobStoreConfig.LockInternalFiles
}
