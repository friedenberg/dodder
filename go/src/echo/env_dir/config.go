package env_dir

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

// TODO move to own package

func MakeConfig(
	hashType markl.HashType,
	funcJoin func(string, ...string) string,
	compression interfaces.BlobCompression,
	encryption interfaces.BlobEncryption,
) Config {
	return Config{
		hashType:    hashType,
		funcJoin:    funcJoin,
		compression: compression,
		encryption:  encryption,
	}
}

var (
	defaultCompressionTypeValue = compression_type.CompressionTypeNone
	defaultEncryptionType       = age.Age{}
	DefaultConfig               = Config{
		hashType:    markl.HashTypeSha256,
		compression: &defaultCompressionTypeValue,
		encryption:  &defaultEncryptionType,
	}

	_ interfaces.BlobIOWrapper = Config{}
)

type Config struct {
	hashType markl.HashType
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
