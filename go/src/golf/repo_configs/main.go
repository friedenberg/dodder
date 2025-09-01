package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type (
	TypedBlob = triple_hyphen_io.TypedBlob[ConfigOverlay]

	DefaultsGetter interface {
		GetDefaults() Defaults
	}

	ConfigOverlay interface {
		DefaultsGetter
		file_extensions.OverlayGetter
		options_print.OverlayGetter
		GetToolOptions() options_tools.Options
	}

	Defaults interface {
		GetDefaultType() ids.Type
		GetDefaultTags() quiter.Slice[ids.Tag]
	}
)

var (
	// _ ConfigOverlay = Config{}
	_ ConfigOverlay = V0{}
	_ ConfigOverlay = V1{}
	_ ConfigOverlay = V2{}
)

func Default(defaultType ids.Type) Config {
	return Config{
		DefaultType:    defaultType,
		DefaultTags:    ids.MakeTagSet(),
		FileExtensions: file_extensions.Default(),
		PrintOptions:   options_print.DefaultOverlay().GetPrintOptionsOverlay(),
		ToolOptions: options_tools.Options{
			Merge: []string{
				"vimdiff",
			},
		},
	}
}

func DefaultOverlay(defaultType ids.Type) TypedBlob {
	return TypedBlob{
		Type: ids.DefaultOrPanic(genres.Config),
		Blob: V2{
			Defaults: DefaultsV1{
				Type: defaultType,
				Tags: make([]ids.Tag, 0),
			},
			PrintOptions:   options_print.DefaultOverlay(),
			FileExtensions: file_extensions.DefaultOverlay(),
			Tools: options_tools.Options{
				Merge: []string{
					"vimdiff",
				},
			},
		},
	}
}
