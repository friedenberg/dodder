package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type DefaultsV0 struct {
	Typ       ids.Type  `toml:"typ"`
	Etiketten []ids.Tag `toml:"etiketten"`
}

func (d DefaultsV0) GetType() ids.Type {
	return d.Typ
}

func (d DefaultsV0) GetTags() quiter.Slice[ids.Tag] {
	return quiter.Slice[ids.Tag](d.Etiketten)
}

type V0 struct {
	Defaults        DefaultsV0                            `toml:"defaults"`
	HiddenEtiketten []ids.Tag                             `toml:"hidden-etiketten"`
	FileExtensions  file_extensions.V0                    `toml:"file-extensions"`
	RemoteScripts   map[string]script_config.RemoteScript `toml:"remote-scripts"`
	Actions         map[string]script_config.ScriptConfig `toml:"actions,omitempty"`
	PrintOptions    options_print.V1                      `toml:"cli-output"`
	Tools           options_tools.Options                 `toml:"tools"`
	Filters         map[string]string                     `toml:"filters"`
}

func (blob V0) GetRepoConfig() Config {
	return blob
}

func (blob *V0) Reset() {
	blob.FileExtensions.Reset()
	blob.Defaults.Typ = ids.Type{}
	blob.Defaults.Etiketten = make([]ids.Tag, 0)
	blob.HiddenEtiketten = make([]ids.Tag, 0)
	blob.RemoteScripts = make(map[string]script_config.RemoteScript)
	blob.Actions = make(map[string]script_config.ScriptConfig)
	blob.PrintOptions = options_print.Default()
	blob.Filters = make(map[string]string)
}

func (blob *V0) ResetWith(b *V0) {
	blob.FileExtensions.Reset()

	blob.Defaults.Typ = b.Defaults.Typ

	blob.Defaults.Etiketten = make([]ids.Tag, len(b.Defaults.Etiketten))
	copy(blob.Defaults.Etiketten, b.Defaults.Etiketten)

	blob.HiddenEtiketten = make([]ids.Tag, len(b.HiddenEtiketten))
	copy(blob.HiddenEtiketten, b.HiddenEtiketten)

	blob.RemoteScripts = b.RemoteScripts
	blob.Actions = b.Actions
	blob.PrintOptions = b.PrintOptions
	blob.Filters = b.Filters
}

func (blob V0) GetFilters() map[string]string {
	return blob.Filters
}

func (blob V0) GetDefaults() Defaults {
	return blob.Defaults
}

func (blob V0) GetFileExtensions() interfaces.FileExtensions {
	return blob.FileExtensions
}

func (blob V0) GetPrintOptions() options_print.Options {
	return blob.PrintOptions.GetPrintOptions()
}
