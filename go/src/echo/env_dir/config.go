package env_dir

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

// TODO move to own package

func MakeConfig(
	envDigest interfaces.EnvBlobId,
	funcJoin func(string, ...string) string,
	compression interfaces.BlobCompression,
	encryption interfaces.BlobEncryption,
) Config {
	return Config{
		envDigest:   envDigest,
		funcJoin:    funcJoin,
		compression: compression,
		encryption:  encryption,
	}
}

var (
	defaultCompressionTypeValue = compression_type.CompressionTypeNone
	defaultEncryptionType       = age.Age{}
	DefaultConfig               = Config{
		envDigest:   sha.Env,
		compression: &defaultCompressionTypeValue,
		encryption:  &defaultEncryptionType,
	}

	_ interfaces.BlobIOWrapper = Config{}
)

type Config struct {
	envDigest interfaces.EnvBlobId
	// TODO replace with path generator interface
	funcJoin    func(string, ...string) string
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
