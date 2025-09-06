package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type TomlV1 struct {
	// TODO add hasher option
	// TODO uncomment when bumping to V1
	// HashBuckets       []int                            `toml:"hash-buckets"`
	BasePath          string                           `toml:"base-path"` // can include env vars
	AgeEncryption     age.Age                          `toml:"age-encryption,omitempty"`
	CompressionType   compression_type.CompressionType `toml:"compression-type"`
	LockInternalFiles bool                             `toml:"lock-internal-files"`
}

var (
	_ ConfigLocalHashBucketed = TomlV1{}
	_ ConfigMutable           = &TomlV1{}
)

func (TomlV1) GetBlobStoreType() string {
	return "local"
}

func (blobStoreConfig *TomlV1) SetFlagSet(flagSet *flags.FlagSet) {
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

func (blobStoreConfig TomlV1) GetBasePath() string {
	return blobStoreConfig.BasePath
}

func (blobStoreConfig TomlV1) GetHashBuckets() []int {
	return []int{2}
}

func (blobStoreConfig TomlV1) GetBlobCompression() interfaces.CommandLineIOWrapper {
	return &blobStoreConfig.CompressionType
}

func (blobStoreConfig TomlV1) GetBlobEncryption() interfaces.CommandLineIOWrapper {
	return &blobStoreConfig.AgeEncryption
}

func (blobStoreConfig TomlV1) GetLockInternalFiles() bool {
	return blobStoreConfig.LockInternalFiles
}

func (blobStoreConfig TomlV1) SupportsMultiHash() bool {
	return true
}

func (blobStoreConfig TomlV1) GetDefaultHashTypeId() string {
	return markl.HashTypeIdSha256
}
