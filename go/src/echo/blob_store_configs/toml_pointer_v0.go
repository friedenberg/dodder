package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type TomlPointerV0 struct {
	ConfigPath string `toml:"config-path"`
}

var (
	_ ConfigPointer = TomlPointerV0{}
	_ ConfigMutable = &TomlPointerV0{}
	_               = registerToml[TomlPointerV0](
		Coder.Blob,
		ids.TypeTomlBlobStoreConfigPointerV0,
	)
)

func (TomlPointerV0) GetBlobStoreType() string {
	return "local-pointer"
}

func (blobStoreConfig *TomlPointerV0) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	flagSet.StringVar(
		&blobStoreConfig.ConfigPath,
		"config-path",
		"",
		"path to another blob store config file",
	)
}

func (blobStoreConfig TomlPointerV0) GetConfigPath() string {
	return blobStoreConfig.ConfigPath
}
