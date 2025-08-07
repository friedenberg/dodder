package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type V2 struct {
	Defaults       DefaultsV1             `toml:"defaults"`
	FileExtensions file_extensions.TOMLV1 `toml:"file-extensions"`
	PrintOptions   options_print.Overlay  `toml:"cli-output"`
	Tools          options_tools.Options  `toml:"tools"`
}

func (config *V2) Reset() {
	config.FileExtensions.Reset()
	config.Defaults.Type = ids.Type{}
	config.Defaults.Tags = make([]ids.Tag, 0)
	config.PrintOptions = options_print.DefaultOverlay().GetPrintOptionsOverlay()
}

func (config *V2) ResetWith(b *V2) {
	config.FileExtensions.Reset()

	config.Defaults.Type = b.Defaults.Type

	config.Defaults.Tags = make([]ids.Tag, len(b.Defaults.Tags))
	copy(config.Defaults.Tags, b.Defaults.Tags)

	config.PrintOptions = b.PrintOptions
}

func (config V2) GetDefaults() Defaults {
	return config.Defaults
}

func (blob V2) GetFileExtensionsOverlay() file_extensions.Overlay {
	return blob.FileExtensions.GetFileExtensionsOverlay()
}

func (config V2) GetPrintOptionsOverlay() options_print.Overlay {
	return config.PrintOptions
}

func (blob V2) GetToolOptions() options_tools.Options {
	return blob.Tools
}
