package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type Config = any

type ConfigMutable interface {
	interfaces.BlobIOWrapper
	interfaces.CommandComponent
}

func Default() ConfigMutable {
	return &TomlV0{
		CompressionType:   compression_type.CompressionTypeDefault,
		LockInternalFiles: true,
	}
}
