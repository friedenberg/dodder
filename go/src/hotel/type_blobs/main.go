package type_blobs

import (
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
)

func Default() (t TomlV1) {
	t = TomlV1{
		FileExtension: "md",
		Formatters:    make(map[string]script_config.WithOutputFormat),
		VimSyntaxType: "markdown",
	}

	return
}

type Blob interface {
	GetFileExtension() string
	GetBinary() bool
	GetVimSyntaxType() string
	WithFormatters
	WithFormatterUTIGropus
	WithStringLuaHooks
}

type WithFormatters interface {
	GetFormatters() map[string]script_config.WithOutputFormat
}

type WithFormatterUTIGropus interface {
	GetFormatterUTIGroups() map[string]UTIGroup
}

// TODO make typed hooks
type WithStringLuaHooks interface {
	GetStringLuaHooks() string
}
