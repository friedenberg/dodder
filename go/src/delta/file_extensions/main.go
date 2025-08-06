package file_extensions

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/equals"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

type (
	OverlayGetter interface {
		GetFileExtensionsOverlay() Overlay
	}

	ConfigGetter interface {
		GetFileExtensions() Config
	}

	Config struct {
		Config   string
		Organize string
		Repo     string
		Tag      string
		Type     string
		Zettel   string
	}

	Overlay struct {
		Config   *string
		Organize *string
		Repo     *string
		Tag      *string
		Type     *string
		Zettel   *string
	}
)

var (
	_ OverlayGetter = TOMLV0{}
	_ OverlayGetter = TOMLV1{}
)

func Default() Config {
	return Config{
		Config:   "konfig",
		Organize: "md",
		Repo:     "repo",
		Tag:      "tag",
		Type:     "type",
		Zettel:   "zettel",
	}
}

func DefaultOverlay() TOMLV1 {
	config := Default()

	return TOMLV1{
		Config:   &config.Config,
		Organize: &config.Organize,
		Repo:     &config.Repo,
		Tag:      &config.Tag,
		Type:     &config.Type,
		Zettel:   &config.Zettel,
	}
}

func MakeDefaultConfig(overlays ...OverlayGetter) Config {
	return MakeConfig(Default(), overlays...)
}

func MakeConfig(base Config, overlays ...OverlayGetter) Config {
	for _, overlayGetter := range overlays {
		overlay := overlayGetter.GetFileExtensionsOverlay()
		equals.SetIfValueNotNil(&base.Config, overlay.Config)
		equals.SetIfValueNotNil(&base.Organize, overlay.Organize)
		equals.SetIfValueNotNil(&base.Repo, overlay.Repo)
		equals.SetIfValueNotNil(&base.Tag, overlay.Tag)
		equals.SetIfValueNotNil(&base.Type, overlay.Type)
		equals.SetIfValueNotNil(&base.Zettel, overlay.Zettel)
	}

	return base
}

func (config Config) GetFileExtensionForGenre(
	getter interfaces.GenreGetter,
) string {
	genre := genres.Must(getter)

	switch genre {
	case genres.Zettel:
		return config.Zettel

	case genres.Type:
		return config.Type

	case genres.Tag:
		return config.Tag

	case genres.Repo:
		return config.Repo

	case genres.Config:
		return config.Config

	default:
		return ""
	}
}
