package file_extensions

type (
	FileExtensionsGetter interface {
		GetFileExtensions() FileExtensions
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
		Zettel   string
		Organize string
		Type     string
		Tag      string
		Repo     string
		Config   string
	}
)

var (
	_ FileExtensions = TOMLV0{}
	_ FileExtensions = TOMLV1{}
)

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

func (config *Config) Reset() {
	config.Zettel = ""
	config.Organize = ""
	config.Type = ""
	config.Tag = ""
	config.Repo = ""
}

func (config *Config) ResetWith(b Config) {
	config.Zettel = b.Zettel
	config.Organize = b.Organize
	config.Type = b.Type
	config.Tag = b.Tag
	config.Repo = b.Repo
}
