package blob_store_config

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type Config interface {
	interfaces.BlobStoreConfigImmutable
}

type ConfigMutable interface {
	Config
	interfaces.CommandComponent
}

func Default() ConfigMutable {
	return &TomlV0{
		CompressionType:   compression_type.CompressionTypeDefault,
		LockInternalFiles: true,
	}
}
