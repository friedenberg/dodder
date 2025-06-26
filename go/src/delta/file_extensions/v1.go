package file_extensions

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

type V1 struct {
	Zettel   string `toml:"zettel"`
	Organize string `toml:"organize"`
	Type     string `toml:"type"`
	Tag      string `toml:"tag"`
	Repo     string `toml:"repo"`
	Config   string `toml:"config"`
}

func (a V1) GetFileExtensionForGenre(
	g1 interfaces.GenreGetter,
) string {
	g := genres.Must(g1)

	switch g {
	case genres.Zettel:
		return a.GetFileExtensionZettel()

	case genres.Type:
		return a.GetFileExtensionType()

	case genres.Tag:
		return a.GetFileExtensionTag()

	case genres.Repo:
		return a.GetFileExtensionRepo()

	case genres.Config:
		return a.GetFileExtensionConfig()

	default:
		return ""
	}
}

func (a V1) GetFileExtensions() interfaces.FileExtensions {
	return a
}

func (a V1) GetFileExtensionZettel() string {
	return a.Zettel
}

func (a V1) GetFileExtensionOrganize() string {
	return a.Organize
}

func (a V1) GetFileExtensionType() string {
	return a.Type
}

func (a V1) GetFileExtensionTag() string {
	return a.Tag
}

func (a V1) GetFileExtensionRepo() string {
	return a.Repo
}

func (a V1) GetFileExtensionConfig() string {
	return a.Config
}

func (a *V1) Reset() {
	a.Zettel = ""
	a.Organize = ""
	a.Type = ""
	a.Tag = ""
	a.Repo = ""
	a.Config = ""
}

func (a *V1) ResetWith(b V1) {
	a.Zettel = b.Zettel
	a.Organize = b.Organize
	a.Type = b.Type
	a.Tag = b.Tag
	a.Repo = b.Repo
	a.Config = b.Config
}
