package type_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/reset"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
)

type TomlV1 struct {
	Binary        bool                                      `toml:"binary,omitempty"`
	FileExtension string                                    `toml:"file-extension,omitempty"`
	ExecCommand   *script_config.ScriptConfig               `toml:"exec-command,omitempty"`
	VimSyntaxType string                                    `toml:"vim-syntax-type"`
	UTIGroups     map[string]UTIGroup                       `toml:"uti-groups"`
	Formatters    map[string]script_config.WithOutputFormat `toml:"formatters,omitempty"`
	// TODO migrate to properly-typed hooks
	Hooks interface{} `toml:"hooks"`
}

func (a *TomlV1) Reset() {
	a.Binary = false
	a.FileExtension = ""
	a.ExecCommand = nil
	a.VimSyntaxType = ""

	a.UTIGroups = reset.Map(a.UTIGroups)
	a.Formatters = reset.Map(a.Formatters)
	a.Hooks = nil
}

func (a *TomlV1) GetBinary() bool {
	return a.Binary
}

func (a *TomlV1) GetFileExtension() string {
	return a.FileExtension
}

func (a *TomlV1) GetVimSyntaxType() string {
	return a.VimSyntaxType
}

func (a *TomlV1) GetFormatters() map[string]script_config.WithOutputFormat {
	return a.Formatters
}

func (a *TomlV1) GetFormatterUTIGroups() map[string]UTIGroup {
	return a.UTIGroups
}

func (a *TomlV1) GetStringLuaHooks() string {
	hooks, _ := a.Hooks.(string)
	return hooks
}
