package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type Config = any

type ConfigMutable interface {
	interfaces.CommandComponent
}

func Default() *TomlV0 {
	return &TomlV0{
		CompressionType:   compression_type.CompressionTypeDefault,
		LockInternalFiles: true,
	}
}
