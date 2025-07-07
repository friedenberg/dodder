package config_immutable

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
)

// TODO move this to a new package, blob_store_config
type BlobStoreType string

const (
	BlobStoreTypeLocal BlobStoreType = "local"
)

type TomlLocalBlobStoreConfigV1 struct {
	AgeEncryption     age.Age         `toml:"age-encryption,omitempty"`
	CompressionType   CompressionType `toml:"compression-type"`
	LockInternalFiles bool            `toml:"lock-internal-files"`
}

func (c *TomlLocalBlobStoreConfigV1) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return c
}

func (c *TomlLocalBlobStoreConfigV1) GetBlobEncryption() interfaces.BlobEncryption {
	return &c.AgeEncryption
}

func (c *TomlLocalBlobStoreConfigV1) GetBlobCompression() interfaces.BlobCompression {
	return &c.CompressionType
}

func (c *TomlLocalBlobStoreConfigV1) GetLockInternalFiles() bool {
	return c.LockInternalFiles
}
