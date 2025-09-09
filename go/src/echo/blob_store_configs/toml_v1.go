package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type TomlV1 struct {
	HashBuckets values.IntSlice `toml:"hash-buckets"`
	BasePath    string          `toml:"base-path"` // can include env vars
	HashTypeId  string          `toml:"hash_type-id"`
	// TODO transform into markl type
	Encryption        markl.Id                         `toml:"encryption"`
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

func (blobStoreConfig *TomlV1) SetFlagSet(
	flagSet interfaces.CommandLineFlagDefinitions,
) {
	blobStoreConfig.CompressionType.SetFlagSet(flagSet)

	flagSet.Var(
		&blobStoreConfig.HashBuckets,
		"hash_buckets",
		"determines hash bucketing directory structure",
	)

	flagSet.StringVar(
		&blobStoreConfig.HashTypeId,
		"hash_type-id",
		markl.HashTypeIdSha256,
		"determines the hash type used for new blobs written to the store",
	)

	flagSet.Func(
		"encryption",
		"add encryption for blobs",
		func(value string) (err error) {
			if files.Exists(value) {
				if err = markl.SetFromPath(
					&blobStoreConfig.Encryption,
					value,
				); err != nil {
					err = errors.Wrapf(err, "Value: %q", value)
					return
				}

				return
			}

			switch value {
			case "none":
				// no-op

			case "", "generate":
				if err = markl.GeneratePrivateKey(
					nil,
					markl.FormatIdMadderPrivateKeyV1,
					markl.TypeIdAgeX25519Sec,
					&blobStoreConfig.Encryption,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

			default:
				if err = blobStoreConfig.Encryption.Set(value); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			return
		},
	)

	flagSet.BoolVar(
		&blobStoreConfig.LockInternalFiles,
		"lock-internal-files",
		blobStoreConfig.LockInternalFiles,
		"",
	)
}

func (blobStoreConfig TomlV1) GetBasePath() string {
	return blobStoreConfig.BasePath
}

func (blobStoreConfig TomlV1) GetHashBuckets() []int {
	return blobStoreConfig.HashBuckets
}

func (blobStoreConfig TomlV1) GetBlobCompression() interfaces.IOWrapper {
	return &blobStoreConfig.CompressionType
}

func (blobStoreConfig TomlV1) GetBlobEncryption() interfaces.MarklId {
	return blobStoreConfig.Encryption
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
