package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

// TODO move to a config_common package
type TomlUriV0 struct {
	Uri values.Uri `toml:"uri"`
}

func (config *TomlUriV0) SetFlagSet(flagSet *flags.FlagSet) {
	flagSet.Var(
		&config.Uri,
		"uri",
		"SFTP server hostname",
	)
}

func (config *TomlUriV0) GetUri() values.Uri {
	return config.Uri
}
