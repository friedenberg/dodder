package config_immutable_io

import (
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

// TODO refactor to common typed blob
type ConfigPrivatedTypedBlob = triple_hyphen_io.TypedBlob[config_immutable.ConfigPrivate]
