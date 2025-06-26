package env_dir

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
)

type Config struct {
	compression       interfaces.BlobCompression
	encryption        interfaces.BlobEncryption
	lockInternalFiles bool
}

func MakeConfigFromImmutableBlobConfig(
	config interfaces.BlobStoreConfigImmutable,
) Config {
	return MakeConfig(
		config.GetBlobCompression(),
		config.GetBlobEncryption(),
		config.GetLockInternalFiles(),
	)
}

func MakeConfig(
	compression interfaces.BlobCompression,
	encryption interfaces.BlobEncryption,
	lockInternalFiles bool,
) Config {
	if encryption == nil {
		encryption = &age.Age{}
	}

	if compression == nil {
		c := config_immutable.CompressionTypeNone
		compression = &c
	}

	return Config{
		compression:       compression,
		encryption:        encryption,
		lockInternalFiles: lockInternalFiles,
	}
}

func (c Config) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return &c
}

func (c *Config) GetBlobCompression() interfaces.BlobCompression {
	return c.compression
}

func (c *Config) GetBlobEncryption() interfaces.BlobEncryption {
	return c.encryption
}

func (c Config) GetLockInternalFiles() bool {
	return c.lockInternalFiles
}
