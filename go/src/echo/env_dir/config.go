package env_dir

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

func MakeConfig(
	compression interfaces.BlobCompression,
	encryption interfaces.BlobEncryption,
	lockInternalFiles bool,
) Config {
	return Config{
		Compression:       compression,
		Encryption:        encryption,
		LockInternalFiles: lockInternalFiles,
	}
}

type Config struct {
	Compression       interfaces.BlobCompression
	Encryption        interfaces.BlobEncryption
	LockInternalFiles bool
}

func (c Config) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return &c
}

func (config Config) GetBlobCompression() interfaces.BlobCompression {
	if config.Compression == nil {
		c := compression_type.CompressionTypeNone
		return &c
	} else {
		return config.Compression
	}
}

func (config Config) GetBlobEncryption() interfaces.BlobEncryption {
	if config.Encryption == nil {
		return &age.Age{}
	} else {
		return config.Encryption
	}
}

func (c Config) GetLockInternalFiles() bool {
	return c.LockInternalFiles
}
