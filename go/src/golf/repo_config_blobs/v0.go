package repo_config_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type V0 struct {
	Defaults        DefaultsV0                            `toml:"defaults"`
	HiddenEtiketten []ids.Tag                             `toml:"hidden-etiketten"`
	FileExtensions  file_extensions.V0                    `toml:"file-extensions"`
	RemoteScripts   map[string]script_config.RemoteScript `toml:"remote-scripts"`
	Actions         map[string]script_config.ScriptConfig `toml:"actions,omitempty"`
	PrintOptions    options_print.V0                      `toml:"cli-output"`
	Tools           options_tools.Options                 `toml:"tools"`
	Filters         map[string]string                     `toml:"filters"`
}

func (a V0) GetBlob() Blob {
	return a
}

func (a *V0) Reset() {
	a.FileExtensions.Reset()
	a.Defaults.Typ = ids.Type{}
	a.Defaults.Etiketten = make([]ids.Tag, 0)
	a.HiddenEtiketten = make([]ids.Tag, 0)
	a.RemoteScripts = make(map[string]script_config.RemoteScript)
	a.Actions = make(map[string]script_config.ScriptConfig)
	a.PrintOptions = options_print.Default()
	a.Filters = make(map[string]string)
}

func (a *V0) ResetWith(b *V0) {
	a.FileExtensions.Reset()

	a.Defaults.Typ = b.Defaults.Typ

	a.Defaults.Etiketten = make([]ids.Tag, len(b.Defaults.Etiketten))
	copy(a.Defaults.Etiketten, b.Defaults.Etiketten)

	a.HiddenEtiketten = make([]ids.Tag, len(b.HiddenEtiketten))
	copy(a.HiddenEtiketten, b.HiddenEtiketten)

	a.RemoteScripts = b.RemoteScripts
	a.Actions = b.Actions
	a.PrintOptions = b.PrintOptions
	a.Filters = b.Filters
}

func (a V0) GetFilters() map[string]string {
	return a.Filters
}

func (a V0) GetDefaults() Defaults {
	return a.Defaults
}

func (a V0) GetFileExtensions() interfaces.FileExtensions {
	return a.FileExtensions
}

func (a V0) GetPrintOptions() options_print.V0 {
	return a.PrintOptions
}
