package env_dir

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

// TODO move to own package

func MakeConfig(
	compression interfaces.BlobCompression,
	encryption interfaces.BlobEncryption,
) Config {
	return Config{
		compression: compression,
		encryption:  encryption,
	}
}

var (
	defaultCompressionTypeValue = compression_type.CompressionTypeNone
	defaultEncryptionType       = age.Age{}
	DefaultConfig               = Config{
		compression: &defaultCompressionTypeValue,
		encryption:  &defaultEncryptionType,
	}

	_ interfaces.BlobIOWrapper = Config{}
)

type Config struct {
	compression interfaces.BlobCompression
	encryption  interfaces.BlobEncryption
}

func (config Config) GetBlobCompression() interfaces.BlobCompression {
	if config.compression == nil {
		return &defaultCompressionTypeValue
	} else {
		return config.compression
	}
}

func (config Config) GetBlobEncryption() interfaces.BlobEncryption {
	if config.encryption == nil {
		return &defaultEncryptionType
	} else {
		return config.encryption
	}
}
