package object_metadata_fmt

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func AddBlobDigestIfNecessary(
	boxContents []string_format_writer.Field,
	digest interfaces.MarklId,
	funcAbbreviate ids.FuncAbbreviateString,
) []string_format_writer.Field {
	value := digest.String()

	if funcAbbreviate != nil {
		abbreviatedDigestString, err := funcAbbreviate(digest)
		if err != nil {
			panic(err)
		}

		if abbreviatedDigestString != "" {
			value = abbreviatedDigestString
		} else {
			ui.Todo("abbreviation func produced empty string")
		}
	}

	if value == "" {
		return boxContents
	}

	field := string_format_writer.Field{
		Value:      "@" + value,
		ColorType:  string_format_writer.ColorTypeHash,
		NoTruncate: true,
	}

	return append(boxContents, field)
}

func AddRepoPubKey(
	boxContents []string_format_writer.Field,
	metadata *object_metadata.Metadata,
) []string_format_writer.Field {
	return addMarklIdIfNotNull(
		boxContents,
		metadata.GetRepoPubKey(),
	)
}

func AddObjectSig(
	boxContents []string_format_writer.Field,
	metadata *object_metadata.Metadata,
) []string_format_writer.Field {
	return append(
		boxContents,
		makeMarklIdField(metadata.GetObjectSig()),
	)
}

func AddMotherSigIfNecessary(
	boxContents []string_format_writer.Field,
	metadata *object_metadata.Metadata,
) []string_format_writer.Field {
	return addMarklIdIfNotNull(
		boxContents,
		metadata.GetMotherObjectSig(),
	)
}

func addMarklIdIfNotNull(
	boxContents []string_format_writer.Field,
	id interfaces.MarklId,
) []string_format_writer.Field {
	if id.IsNull() {
		return boxContents
	}

	return addMarklId(boxContents, id)
}

func addMarklId(
	boxContents []string_format_writer.Field,
	id interfaces.MarklId,
) []string_format_writer.Field {
	return append(
		boxContents,
		makeMarklIdField(id),
	)
}

func makeMarklIdField(
	id interfaces.MarklId,
) string_format_writer.Field {
	if id.GetPurpose() == "" {
		panic(fmt.Sprintf("empty format for markl id: %q", id))
	}

	return string_format_writer.Field{
		Key:        id.GetPurpose(),
		Separator:  '@',
		Value:      id.String(),
		NoTruncate: true,
		ColorType:  string_format_writer.ColorTypeHash,
	}
}
