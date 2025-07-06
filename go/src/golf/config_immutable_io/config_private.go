package config_immutable_io

import (
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// TODO refactor to common typed blob
type ConfigPrivatedTypedBlob struct {
	ids.Type
	ImmutableConfig config_immutable.ConfigPrivate
}

func (typedBlob *ConfigPrivatedTypedBlob) GetType() ids.Type {
	return typedBlob.Type
}

func (typedBlob *ConfigPrivatedTypedBlob) SetType(t ids.Type) {
	typedBlob.Type = t
}
