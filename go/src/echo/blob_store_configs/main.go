package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
)

type Config = any

type ConfigMutable interface {
	interfaces.CommandComponent
}

type ConfigLocalGitLikeBucketed interface {
	interfaces.BlobIOWrapper
	GetLockInternalFiles() bool
}

var (
	_ ConfigLocalGitLikeBucketed = &TomlV0{}
	_ ConfigMutable              = &TomlV0{}
	_ ConfigMutable              = &TomlSftpV0{}
)

func Default() *TomlV0 {
	return &TomlV0{
		CompressionType:   compression_type.CompressionTypeDefault,
		LockInternalFiles: true,
	}
}
