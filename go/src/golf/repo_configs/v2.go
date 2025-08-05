package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type V2 struct {
	Defaults       DefaultsV1            `toml:"defaults"`
	FileExtensions file_extensions.V1    `toml:"file-extensions"`
	PrintOptions   options_print.Options `toml:"cli-output"`
	Tools          options_tools.Options `toml:"tools"`
}

func (config *V2) Reset() {
	config.FileExtensions.Reset()
	config.Defaults.Type = ids.Type{}
	config.Defaults.Tags = make([]ids.Tag, 0)
	config.PrintOptions = options_print.Default().GetPrintOptions()
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

func (config V2) GetFileExtensions() interfaces.FileExtensions {
	return config.FileExtensions
}

func (config V2) GetPrintOptions() options_print.Options {
	return config.PrintOptions
}
