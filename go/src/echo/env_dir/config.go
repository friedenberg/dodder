package env_dir

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
)

// TODO move to own blob store configs package

func MakeConfig(
	hashFormat interfaces.FormatHash,
	funcJoin func(string, ...string) string,
	compression interfaces.IOWrapper,
	encryption interfaces.MarklId,
) Config {
	var ioWrapper interfaces.IOWrapper = defaultEncryptionIOWrapper

	if encryption != nil {
		var err error
		ioWrapper, err = encryption.GetIOWrapper()
		errors.PanicIfError(err)
	}

	return Config{
		hashFormat:  hashFormat,
		funcJoin:    funcJoin,
		compression: compression,
		encryption:  ioWrapper,
	}
}

var (
	defaultCompressionTypeValue = compression_type.CompressionTypeNone
	defaultEncryptionIOWrapper  = files.NopeIOWrapper{}
	DefaultConfig               = Config{
		hashFormat:  blob_store_configs.DefaultHashType,
		compression: &defaultCompressionTypeValue,
		encryption:  &defaultEncryptionIOWrapper,
	}
)

type Config struct {
	hashFormat interfaces.FormatHash
	// TODO replace with path generator interface
	funcJoin    func(string, ...string) string
	compression interfaces.IOWrapper
	encryption  interfaces.IOWrapper
}

func (config Config) GetBlobCompression() interfaces.IOWrapper {
	if config.compression == nil {
		return &defaultCompressionTypeValue
	} else {
		return config.compression
	}
}

func (config Config) GetBlobEncryption() interfaces.IOWrapper {
	if config.encryption == nil {
		return defaultEncryptionIOWrapper
	} else {
		return config.encryption
	}
}
