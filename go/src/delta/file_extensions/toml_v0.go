package file_extensions

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type TOMLV0 struct {
	Zettel   string `toml:"zettel"`
	Organize string `toml:"organize"`
	Type     string `toml:"typ"`
	Tag      string `toml:"etikett"`
	Repo     string `toml:"kasten"`
}

func (config TOMLV0) GetFileExtensions() interfaces.FileExtensions {
	return config
}

func (config TOMLV0) GetFileExtensionZettel() string {
	return config.Zettel
}

func (config TOMLV0) GetFileExtensionOrganize() string {
	return config.Organize
}

func (config TOMLV0) GetFileExtensionType() string {
	return config.Type
}

func (config TOMLV0) GetFileExtensionTag() string {
	return config.Tag
}

func (config TOMLV0) GetFileExtensionRepo() string {
	return config.Repo
}

func (config TOMLV0) GetFileExtensionConfig() string {
	return "konfig"
}

func (config *TOMLV0) Reset() {
	config.Zettel = ""
	config.Organize = ""
	config.Type = ""
	config.Tag = ""
	config.Repo = ""
}

func (config *TOMLV0) ResetWith(b TOMLV0) {
	config.Zettel = b.Zettel
	config.Organize = b.Organize
	config.Type = b.Type
	config.Tag = b.Tag
	config.Repo = b.Repo
}
