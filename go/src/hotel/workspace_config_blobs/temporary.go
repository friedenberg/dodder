package workspace_config_blobs

import "code.linenisgreat.com/dodder/go/src/golf/repo_config"

type Temporary struct {
	Defaults repo_config.DefaultsV1OmitEmpty `toml:"defaults,omitempty"`
}

func (Temporary) temporaryWorkspace() {}

func (blob Temporary) GetDefaults() repo_config.Defaults {
	return blob.Defaults
}
