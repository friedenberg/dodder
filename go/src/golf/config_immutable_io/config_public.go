package config_immutable_io

import (
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type ConfigLoadedPublic struct {
	ids.Type
	ImmutableConfig          config_immutable.ConfigPublic // TODO restructure this to avoid this extra call
	BlobStoreImmutableConfig env_dir.Config // TODO extricate from env_dir
}

func (c *ConfigLoadedPublic) GetType() ids.Type {
	return c.Type
}

func (c *ConfigLoadedPublic) SetType(t ids.Type) {
	c.Type = t
}
