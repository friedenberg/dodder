package file_extensions

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type TOMLV1 struct {
	Zettel   string `toml:"zettel"`
	Organize string `toml:"organize"`
	Type     string `toml:"type"`
	Tag      string `toml:"tag"`
	Repo     string `toml:"repo"`
	Config   string `toml:"config"`
}

func (config TOMLV1) GetFileExtensions() interfaces.FileExtensions {
	return config
}

func (config TOMLV1) GetFileExtensionZettel() string {
	return config.Zettel
}

func (config TOMLV1) GetFileExtensionOrganize() string {
	return config.Organize
}

func (config TOMLV1) GetFileExtensionType() string {
	return config.Type
}

func (config TOMLV1) GetFileExtensionTag() string {
	return config.Tag
}

func (config TOMLV1) GetFileExtensionRepo() string {
	return config.Repo
}

func (config TOMLV1) GetFileExtensionConfig() string {
	return config.Config
}

func (config *TOMLV1) Reset() {
	config.Zettel = ""
	config.Organize = ""
	config.Type = ""
	config.Tag = ""
	config.Repo = ""
	config.Config = ""
}

func (config *TOMLV1) ResetWith(b TOMLV1) {
	config.Zettel = b.Zettel
	config.Organize = b.Organize
	config.Type = b.Type
	config.Tag = b.Tag
	config.Repo = b.Repo
	config.Config = b.Config
}
