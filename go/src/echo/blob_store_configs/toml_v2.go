package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type TomlV2 struct {
	HashBuckets values.IntSlice `toml:"hash_buckets"`
	BasePath    string          `toml:"base_path"` // can include env vars
	HashTypeId  string          `toml:"hash_type-id"`

	// cannot use `omitempty`, as markl.Id's empty value equals its non-empty
	// value due to unexported fields
	Encryption markl.Id `toml:"encryption"`

	CompressionType   compression_type.CompressionType `toml:"compression-type"`
	LockInternalFiles bool                             `toml:"lock-internal-files"`
}

var (
	_ ConfigLocalHashBucketed = TomlV2{}
	_ ConfigLocalMutable      = &TomlV2{}
)

func (TomlV2) GetBlobStoreType() string {
	return "local"
}

func (blobStoreConfig *TomlV2) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	blobStoreConfig.CompressionType.SetFlagDefinitions(flagSet)

	blobStoreConfig.HashBuckets = DefaultHashBuckets

	flagSet.Var(
		&blobStoreConfig.HashBuckets,
		"hash_buckets",
		"determines hash bucketing directory structure",
	)

	flagSet.StringVar(
		&blobStoreConfig.HashTypeId,
		"hash_type-id",
		markl.FormatIdHashBlake2b256,
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
					return err
				}

				return err
			}

			switch value {
			case "none":
				// no-op

			case "", "generate":
				if err = blobStoreConfig.Encryption.GeneratePrivateKey(
					nil,
					markl.FormatIdSecAgeX25519,
					markl.PurposeMadderPrivateKeyV1,
				); err != nil {
					err = errors.Wrap(err)
					return err
				}

			default:
				if err = blobStoreConfig.Encryption.Set(value); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}

			return err
		},
	)

	flagSet.BoolVar(
		&blobStoreConfig.LockInternalFiles,
		"lock-internal-files",
		blobStoreConfig.LockInternalFiles,
		"",
	)
}

func (blobStoreConfig TomlV2) getBasePath() string {
	return blobStoreConfig.BasePath
}

func (blobStoreConfig TomlV2) GetHashBuckets() []int {
	return blobStoreConfig.HashBuckets
}

func (blobStoreConfig TomlV2) GetBlobCompression() interfaces.IOWrapper {
	return &blobStoreConfig.CompressionType
}

func (blobStoreConfig TomlV2) GetBlobEncryption() interfaces.MarklId {
	return blobStoreConfig.Encryption
}

func (blobStoreConfig TomlV2) GetLockInternalFiles() bool {
	return blobStoreConfig.LockInternalFiles
}

func (blobStoreConfig TomlV2) SupportsMultiHash() bool {
	return true
}

func (blobStoreConfig TomlV2) GetDefaultHashTypeId() string {
	return blobStoreConfig.HashTypeId
}

func (blobStoreConfig *TomlV2) setBasePath(value string) {
	blobStoreConfig.BasePath = value
}
