package config_immutable_io

import (
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// TODO refactor to common typed blob
type ConfigPublicTypedBlob struct {
	ids.Type
	ImmutableConfig config_immutable.ConfigPublic // TODO restructure this to avoid this extra call
}

func (typedBlob *ConfigPublicTypedBlob) GetType() ids.Type {
	return typedBlob.Type
}
