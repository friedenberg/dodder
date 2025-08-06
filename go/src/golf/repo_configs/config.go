package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type Config struct {
	DefaultType    ids.Type
	DefaultTags    ids.TagSet
	FileExtensions interfaces.FileExtensions
	PrintOptions   options_print.Options
	ToolOptions    options_tools.Options
}

type ConfigOverlay2 struct {
	DefaultType *ids.Type
	DefaultTags ids.TagSet
}

func MakeConfigFromOverlays(base Config, overlays ...ConfigOverlay) Config {
	return Config{}
}

func (config Config) GetToolOptions() options_tools.Options {
	return config.ToolOptions
}
