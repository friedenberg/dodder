package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type TomlV1 struct {
	HashBuckets values.IntSlice `toml:"hash-buckets"`
	BasePath    string          `toml:"base-path"` // can include env vars
	HashTypeId  string          `toml:"hash_type-id"`

	// cannot use `omitempty`, as markl.Id's empty value equals its non-empty
	// value due to unexported fields
	Encryption markl.Id `toml:"encryption"`

	CompressionType   compression_type.CompressionType `toml:"compression-type"`
	LockInternalFiles bool                             `toml:"lock-internal-files"`
}

var (
	_ ConfigLocalHashBucketed = TomlV1{}
	_ ConfigUpgradeable       = TomlV1{}
	_ ConfigLocalMutable      = &TomlV1{}
)

func (TomlV1) GetBlobStoreType() string {
	return "local"
}

func (blobStoreConfig *TomlV1) SetFlagDefinitions(
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

func (blobStoreConfig TomlV1) getBasePath() string {
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
	return blobStoreConfig.HashTypeId
}

func (blobStoreConfig *TomlV1) setBasePath(value string) {
	blobStoreConfig.BasePath = value
}

func (blobStoreConfig TomlV1) Upgrade() (Config, ids.Type) {
	upgraded := &TomlV2{
		HashBuckets:       blobStoreConfig.HashBuckets,
		BasePath:          blobStoreConfig.BasePath,
		HashTypeId:        markl.FormatIdHashSha256,
		CompressionType:   blobStoreConfig.CompressionType,
		LockInternalFiles: blobStoreConfig.LockInternalFiles,
	}

	upgraded.Encryption.ResetWithMarklId(blobStoreConfig.Encryption)

	return upgraded, ids.GetOrPanic(ids.TypeTomlBlobStoreConfigV2).Type
}
