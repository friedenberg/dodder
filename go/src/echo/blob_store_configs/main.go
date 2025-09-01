package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type (
	Config        = interfaces.BlobStoreConfig
	ConfigMutable interface {
		interfaces.BlobStoreConfig
		flags.CommandComponentWriter
	}

	ConfigLocalSingleHashBucketed interface {
		Config
		interfaces.BlobIOWrapper
		GetHashBuckets() []int
		GetLockInternalFiles() bool
		GetHashTypeId() string
	}

	ConfigLocalMultiHashBucketed interface {
		Config
		interfaces.BlobIOWrapper
		GetHashBuckets() []int
		GetLockInternalFiles() bool
	}

	// TODO add config interface for local file stores
	ConfigLocalHashBucketed interface {
		Config
		interfaces.BlobIOWrapper
		GetHashBuckets() []int
		GetLockInternalFiles() bool
	}

	ConfigSFTPRemotePath interface {
		Config
		GetRemotePath() string
	}

	ConfigSFTPUri interface {
		ConfigSFTPRemotePath

		GetUri() values.Uri
	}

	ConfigSFTPConfigExplicit interface {
		ConfigSFTPRemotePath

		GetHost() string
		GetPort() int
		GetUser() string
		GetPassword() string
		GetPrivateKeyPath() string
	}

	TypedConfig        = triple_hyphen_io.TypedBlob[Config]
	TypedMutableConfig = triple_hyphen_io.TypedBlob[ConfigMutable]
)

var (
	_ ConfigLocalHashBucketed = &TomlV0{}
	_ ConfigSFTPRemotePath    = &TomlSFTPV0{}
	_ ConfigSFTPRemotePath    = &TomlSFTPViaSSHConfigV0{}
	_ ConfigMutable           = &TomlV0{}
	_ ConfigMutable           = &TomlSFTPV0{}
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
