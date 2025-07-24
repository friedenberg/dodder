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
	GetMimeType() string
	GetVimSyntaxType() string
	WithFormatters
	WithFormatterUTIGroups
	WithStringLuaHooks
}

var (
	_ Blob = &TomlV0{}
	_ Blob = &TomlV1{}
)

type WithFormatters interface {
	GetFormatters() map[string]script_config.WithOutputFormat
}

type WithFormatterUTIGroups interface {
	GetFormatterUTIGroups() map[string]UTIGroup
}

// TODO make typed hooks
type WithStringLuaHooks interface {
	GetStringLuaHooks() string
}
