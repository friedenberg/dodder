package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type (
	Config interface {
		GetBlobStoreType() string
	}

	ConfigMutable interface {
		Config
		interfaces.CommandComponent
	}

	ConfigLocalGitLikeBucketed interface {
		interfaces.BlobIOWrapper
		GetLockInternalFiles() bool
	}

	TypedConfig        = triple_hyphen_io.TypedBlob[Config]
	TypedMutableConfig = triple_hyphen_io.TypedBlob[ConfigMutable]
)

var (
	_ ConfigLocalGitLikeBucketed = &TomlV0{}
	_ ConfigMutable              = &TomlV0{}
	_ ConfigMutable              = &TomlSftpV0{}
)

func Default() *TypedMutableConfig {
	return &TypedMutableConfig{
		Type: ids.GetOrPanic(ids.TypeTomlBlobStoreConfigV0).Type,
		Blob: &TomlV0{
			CompressionType:   compression_type.CompressionTypeDefault,
			LockInternalFiles: true,
		},
	}
}
