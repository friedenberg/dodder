package blob_store_configs

import (
	"flag"
)

type TomlSFTPViaSSHConfigV0 struct {
	TomlUriV0
}

func (*TomlSFTPViaSSHConfigV0) GetBlobStoreType() string {
	return "sftp"
}

func (config *TomlSFTPViaSSHConfigV0) SetFlagSet(
	flagSet *flag.FlagSet,
) {
	config.TomlUriV0.SetFlagSet(flagSet)
}

func (config *TomlSFTPViaSSHConfigV0) GetRemotePath() string {
	uri := config.TomlUriV0.GetUri()
	return uri.GetUrl().Path
}
