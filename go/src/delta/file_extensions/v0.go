package file_extensions

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

type V0 struct {
	Zettel   string `toml:"zettel"`
	Organize string `toml:"organize"`
	Type     string `toml:"typ"`
	Tag      string `toml:"etikett"`
	Repo     string `toml:"kasten"`
}

func (a V0) GetFileExtensionForGenre(
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

	default:
		return ""
	}
}

func (a V0) GetFileExtensions() interfaces.FileExtensions {
	return a
}

func (a V0) GetFileExtensionZettel() string {
	return a.Zettel
}

func (a V0) GetFileExtensionOrganize() string {
	return a.Organize
}

func (a V0) GetFileExtensionType() string {
	return a.Type
}

func (a V0) GetFileExtensionTag() string {
	return a.Tag
}

func (a V0) GetFileExtensionRepo() string {
	return a.Repo
}

func (a V0) GetFileExtensionConfig() string {
	return "konfig"
}

func (a *V0) Reset() {
	a.Zettel = ""
	a.Organize = ""
	a.Type = ""
	a.Tag = ""
	a.Repo = ""
}

func (a *V0) ResetWith(b V0) {
	a.Zettel = b.Zettel
	a.Organize = b.Organize
	a.Type = b.Type
	a.Tag = b.Tag
	a.Repo = b.Repo
}
