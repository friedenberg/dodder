package config_mutable_blobs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

const (
	TypeV0 = builtin_types.ConfigTypeTomlV0
	TypeV1 = builtin_types.ConfigTypeTomlV1
)

type (
	Getter interface {
		GetMutableConfig() Blob
	}

	Blob interface {
		interfaces.MutableStoredConfig
		GetBlob() Blob
		GetDefaults() Defaults
		GetFileExtensions() interfaces.FileExtensions
		GetPrintOptions() options_print.V0
	}

	Defaults interface {
		GetType() ids.Type
		GetTags() quiter.Slice[ids.Tag]
	}
)

func Default(defaultTyp ids.Type) ids.TypedBlob[Blob] {
	return ids.TypedBlob[Blob]{
		Type: builtin_types.DefaultOrPanic(genres.Config),
		Blob: V1{
			Defaults: DefaultsV1{
				Type: defaultTyp,
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
