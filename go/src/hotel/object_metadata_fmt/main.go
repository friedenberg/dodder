package object_metadata_fmt

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/collections_slice"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func AddBlobDigestIfNecessary(
	boxContents collections_slice.Slice[string_format_writer.Field],
	digest interfaces.MarklId,
	funcAbbreviate ids.FuncAbbreviateString,
) {
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
		return
	}

	field := string_format_writer.Field{
		Value:      "@" + value,
		ColorType:  string_format_writer.ColorTypeHash,
		NoTruncate: true,
	}

	boxContents.Append(field)
}

func AddRepoPubKey(
	boxContents collections_slice.Slice[string_format_writer.Field],
	metadata *object_metadata.Metadata,
) {
	addMarklIdIfNotNull(
		boxContents,
		metadata.GetRepoPubKey(),
	)
}

func AddObjectSig(
	boxContents collections_slice.Slice[string_format_writer.Field],
	metadata *object_metadata.Metadata,
) {
	boxContents.Append(
		makeMarklIdField(metadata.GetObjectSig()),
	)
}

func AddMotherSigIfNecessary(
	boxContents collections_slice.Slice[string_format_writer.Field],
	metadata *object_metadata.Metadata,
) {
	addMarklIdIfNotNull(
		boxContents,
		metadata.GetMotherObjectSig(),
	)
}

func AddReferencedObject(
	boxContents collections_slice.Slice[string_format_writer.Field],
	metadata *object_metadata.Metadata,
) {
	boxContents.Append(
		makeMarklIdField(metadata.GetObjectSig()),
	)
}

func addMarklIdIfNotNull(
	boxContents collections_slice.Slice[string_format_writer.Field],
	id interfaces.MarklId,
) {
	if id.IsNull() {
		return
	}

	addMarklId(boxContents, id)
}

func addMarklId(
	boxContents collections_slice.Slice[string_format_writer.Field],
	id interfaces.MarklId,
) {
	boxContents.Append(
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
