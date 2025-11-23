package typed_blob_store

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/object_metadata_fmt_triple_hyphen"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func MakeTextFormatter(
	envRepo env_repo.Env,
	options checkout_options.TextFormatterOptions,
	inlineTypeChecker ids.InlineTypeChecker,
	checkoutMode checkout_mode.Mode,
) textFormatter {
	return MakeTextFormatterWithBlobFormatter(
		envRepo,
		options,
		inlineTypeChecker,
		nil,
		checkoutMode,
	)
}

func MakeTextFormatterWithBlobFormatter(
	envRepo env_repo.Env,
	options checkout_options.TextFormatterOptions,
	inlineTypeChecker ids.InlineTypeChecker,
	formatter script_config.RemoteScript,
	checkoutMode checkout_mode.Mode,
) textFormatter {
	return textFormatter{
		options:           options,
		InlineTypeChecker: inlineTypeChecker,
		TextFormatterFamily: object_metadata_fmt_triple_hyphen.MakeTextFormatterFamily(
			object_metadata_fmt_triple_hyphen.Dependencies{
				EnvDir:        envRepo,
				BlobStore:     envRepo.GetDefaultBlobStore(),
				BlobFormatter: formatter,
			},
		),
		checkoutMode: checkoutMode,
	}
}

type textFormatter struct {
	ids.InlineTypeChecker
	options checkout_options.TextFormatterOptions
	object_metadata_fmt_triple_hyphen.TextFormatterFamily
	checkoutMode checkout_mode.Mode
}

func (formatter textFormatter) EncodeStringTo(
	object *sku.Transacted,
	writer io.Writer,
) (n int64, err error) {
	context := object_metadata_fmt_triple_hyphen.TextFormatterContext{
		PersistentFormatterContext: object,
		TextFormatterOptions:       formatter.options,
	}

	switch {
	case formatter.checkoutMode.IsMetadataOnly():
		n, err = formatter.MetadataOnly.FormatMetadata(writer, context)

	default:
		if genres.Config.EqualsGenre(object.GetGenre()) {
			n, err = formatter.BlobOnly.FormatMetadata(writer, context)
		} else if formatter.InlineTypeChecker.IsInlineType(object.GetType()) {
			n, err = formatter.InlineBlob.FormatMetadata(writer, context)
		} else {
			n, err = formatter.MetadataOnly.FormatMetadata(writer, context)
		}
	}

	return n, err
}

func (tf textFormatter) WriteStringFormatWithMode(
	w io.Writer,
	sk *sku.Transacted,
	mode checkout_mode.Mode,
) (n int64, err error) {
	ctx := object_metadata_fmt_triple_hyphen.TextFormatterContext{
		PersistentFormatterContext: sk,
		TextFormatterOptions:       tf.options,
	}

	if genres.Config.EqualsGenre(sk.GetGenre()) ||
		mode.IsBlobOnly() {
		n, err = tf.BlobOnly.FormatMetadata(w, ctx)
	} else if tf.InlineTypeChecker.IsInlineType(sk.GetType()) {
		n, err = tf.InlineBlob.FormatMetadata(w, ctx)
	} else {
		n, err = tf.MetadataOnly.FormatMetadata(w, ctx)
	}

	return n, err
}
