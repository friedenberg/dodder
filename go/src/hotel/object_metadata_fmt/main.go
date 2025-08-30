package object_metadata_fmt

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
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
