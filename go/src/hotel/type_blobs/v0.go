package type_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/reset"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
)

type V0 struct {
	InlineBlob    bool                        `toml:"inline-akte,omitempty"`
	Archived      bool                        `toml:"archived,omitempty"`
	FileExtension string                      `toml:"file-extension,omitempty"`
	ExecCommand   *script_config.ScriptConfig `toml:"exec-command,omitempty"`
	VimSyntaxType string                      `toml:"vim-syntax-type"`
	// TODO-P4 rename to uti-groups
	FormatterUTIGroups map[string]UTIGroup                       `toml:"formatter-uti-groups"`
	Formatters         map[string]script_config.WithOutputFormat `toml:"formatters,omitempty"`
	Actions            map[string]script_config.ScriptConfig     `toml:"actions,omitempty"`
	Hooks              interface{}                               `toml:"hooks"`
}

func (a *V0) Reset() {
	a.Archived = false
	a.InlineBlob = false
	a.FileExtension = ""
	a.ExecCommand = nil
	a.VimSyntaxType = ""

	a.FormatterUTIGroups = reset.Map(a.FormatterUTIGroups)
	a.Formatters = reset.Map(a.Formatters)
	a.Actions = reset.Map(a.Actions)
	a.Hooks = nil
}

func (a *V0) GetBinary() bool {
	return !a.InlineBlob
}

func (a *V0) GetFileExtension() string {
	return a.FileExtension
}

func (a *V0) GetVimSyntaxType() string {
	return a.VimSyntaxType
}

func (a *V0) GetFormatters() map[string]script_config.WithOutputFormat {
	return a.Formatters
}

func (a *V0) GetFormatterUTIGroups() map[string]UTIGroup {
	return a.FormatterUTIGroups
}

func (a *V0) GetStringLuaHooks() string {
	hooks, _ := a.Hooks.(string)
	return hooks
}
