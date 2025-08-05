package workspace_config_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
)

type (
	TypedConfig = triple_hyphen_io.TypedBlob[Config]

	Config interface {
		GetDefaults() repo_configs.Defaults
	}

	ConfigTemporary interface {
		Config
		temporaryWorkspace()
	}

	ConfigWithDefaultQueryString interface {
		Config
		GetDefaultQueryString() string
	}

	ConfigWithDryRun interface {
		Config
		interfaces.ConfigDryRunGetter
	}
)

var (
	_ ConfigWithDefaultQueryString = V0{}
	_ ConfigTemporary              = Temporary{}
)
