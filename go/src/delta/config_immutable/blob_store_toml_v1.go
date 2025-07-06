package config_immutable

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
)

type BlobStoreTomlV1 struct {
	AgeEncryption     age.Age         `toml:"age-encryption,omitempty"`
	CompressionType   CompressionType `toml:"compression-type"`
	LockInternalFiles bool            `toml:"lock-internal-files"`
}

func (blobStoreConfig *BlobStoreTomlV1) SetFlagSet(f *flag.FlagSet) {
	blobStoreConfig.CompressionType.SetFlagSet(f)

	f.BoolVar(
		&blobStoreConfig.LockInternalFiles,
		"lock-internal-files",
		blobStoreConfig.LockInternalFiles,
		"",
	)

	f.Var(&blobStoreConfig.AgeEncryption, "age-identity", "add an age identity")
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

func (blobStoreConfig *BlobStoreTomlV1) GetAgeEncryption() *age.Age {
	return &blobStoreConfig.AgeEncryption
}

func (blobStoreConfig *BlobStoreTomlV1) GetLockInternalFiles() bool {
	return blobStoreConfig.LockInternalFiles
}
