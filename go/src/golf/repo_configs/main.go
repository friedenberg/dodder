package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/bravo/blob_store_id"
	"code.linenisgreat.com/dodder/go/src/bravo/collections_slice"
	"code.linenisgreat.com/dodder/go/src/bravo/options_tools"
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

	ConfigOverlay2 interface {
		ConfigOverlay
		GetDefaultBlobStoreId() blob_store_id.Id
	}

	Defaults interface {
		GetDefaultType() ids.Type
		GetDefaultTags() collections_slice.Slice[ids.Tag]
	}
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

func DefaultOverlay(
	defaultBlobStoreId blob_store_id.Id,
	defaultType ids.Type,
) TypedBlob {
	return TypedBlob{
		Type: ids.DefaultOrPanic(genres.Config),
		Blob: V2{
			DefaultBlobStoreId: defaultBlobStoreId,
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

func GetDefaultBlobStoreId(
	config ConfigOverlay,
	otherwise blob_store_id.Id,
) blob_store_id.Id {
	if config, ok := config.(ConfigOverlay2); ok {
		return config.GetDefaultBlobStoreId()
	} else {
		return otherwise
	}
}
