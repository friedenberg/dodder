package file_extensions

import "code.linenisgreat.com/dodder/go/src/bravo/equals"

type (
	OverlayGetter interface {
		GetFileExtensionsOverlay() Overlay
	}

	ConfigGetter interface {
		GetFileExtensions() Config
	}

	FileExtensions interface {
		GetFileExtensionZettel() string
		GetFileExtensionOrganize() string
		GetFileExtensionType() string
		GetFileExtensionTag() string
		GetFileExtensionRepo() string
		GetFileExtensionConfig() string
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

func (config Config) GetFileExtensions() FileExtensions {
	return config
}

func (config Config) GetFileExtensionZettel() string {
	return config.Zettel
}

func (config Config) GetFileExtensionOrganize() string {
	return config.Organize
}

func (config Config) GetFileExtensionType() string {
	return config.Type
}

func (config Config) GetFileExtensionTag() string {
	return config.Tag
}

func (config Config) GetFileExtensionRepo() string {
	return config.Repo
}

func (config Config) GetFileExtensionConfig() string {
	return "konfig"
}
