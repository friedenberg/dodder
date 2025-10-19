package blob_store_configs

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

const DefaultHashTypeId = markl.FormatIdHashSha256

var DefaultHashType markl.FormatHash = markl.FormatHashSha256

type (
	Config = interface {
		GetBlobStoreType() string
	}

	ConfigUpgradeable interface {
		Config
		Upgrade() (Config, ids.Type)
	}

	ConfigMutable interface {
		Config
		interfaces.CommandComponentWriter
	}

	ConfigHashType interface {
		SupportsMultiHash() bool
		GetDefaultHashTypeId() string
	}

	configLocal interface {
		Config
		getBasePath() string
	}

	// TODO make private and use public static setter pattern
	ConfigLocalMutable interface {
		configLocal
		SetBasePath(string)
	}

	ConfigLocalHashBucketed interface {
		configLocal
		ConfigHashType
		interfaces.BlobIOWrapper
		GetHashBuckets() []int
		GetLockInternalFiles() bool
	}

	ConfigPointer interface {
		Config
		GetConfigPath() string
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
	_ ConfigSFTPRemotePath = &TomlSFTPV0{}
	_ ConfigSFTPRemotePath = &TomlSFTPViaSSHConfigV0{}
	_ ConfigMutable        = &TomlV0{}
	_ ConfigMutable        = &TomlSFTPV0{}
)

var DefaultHashBuckets []int = []int{2}

type DefaultType = TomlV2

func Default() *TypedMutableConfig {
	return &TypedMutableConfig{
		Type: ids.GetOrPanic(ids.TypeTomlBlobStoreConfigVCurrent).Type,
		Blob: &DefaultType{
			HashBuckets:       DefaultHashBuckets,
			HashTypeId:        markl.FormatIdHashBlake2b256,
			CompressionType:   compression_type.CompressionTypeDefault,
			LockInternalFiles: true,
		},
	}
}

func GetBasePath(config Config) (string, bool) {
	configLocal, ok := config.(configLocal)

	if !ok {
		return "", false
	}

	return os.ExpandEnv(configLocal.getBasePath()), true
}

// func SetBasePath(config Config, basePath string) error {
// 	configLocalMutable, ok := config.(ConfigLocalMutable)

// 	if !ok {
// 		return errors.Errorf("kk
// 	}

// 	return os.ExpandEnv(configLocalMutable.getBasePath()), true
// }
