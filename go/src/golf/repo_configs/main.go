package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type (
	TypedBlob = triple_hyphen_io.TypedBlob[Config]

	Config interface {
		GetDefaults() Defaults
		GetFileExtensions() interfaces.FileExtensions
		GetPrintOptions() options_print.Options
	}

	Defaults interface {
		GetType() ids.Type
		GetTags() quiter.Slice[ids.Tag]
	}
)

var (
	_ Config = V0{}
	_ Config = V1{}
	_ Config = V2{}
)

func Default(defaultType ids.Type) TypedBlob {
	return TypedBlob{
		Type: ids.DefaultOrPanic(genres.Config),
		Blob: V2{
			Defaults: DefaultsV1{
				Type: defaultType,
				Tags: make([]ids.Tag, 0),
			},
			FileExtensions: file_extensions.V1{
				Type:     "type",
				Zettel:   "zettel",
				Organize: "md",
				Tag:      "tag",
				Repo:     "repo",
				Config:   "konfig",
			},
			Tools: options_tools.Options{
				Merge: []string{
					"vimdiff",
				},
			},
		},
	}
}
