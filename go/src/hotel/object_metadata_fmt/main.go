package object_metadata_fmt

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

func AddBlobDigestIfNecessary(
	boxContents []string_format_writer.Field,
	digest interfaces.MarklId,
	abbr ids.FuncAbbreviateString,
) []string_format_writer.Field {
	value := digest.String()

	if abbr != nil {
		abbreviatedDigestString, err := abbr(digest)
		if err != nil {
			panic(err)
		}

		if abbreviatedDigestString != "" {
			value = abbreviatedDigestString
		} else {
			ui.Todo("abbreviate sha produced empty string")
		}
	}

	if value == "" {
		return boxContents
	}

	var field string_format_writer.Field

	if digest.GetType() == markl.HRPObjectBlobDigestSha256V0 {
		field = string_format_writer.Field{
			Value:     "@" + value,
			ColorType: string_format_writer.ColorTypeHash,
		}
	} else {
		field = string_format_writer.Field{
			Key:       digest.GetType(),
			Value:     "@" + value,
			ColorType: string_format_writer.ColorTypeHash,
		}
	}

	return append(boxContents, field)
}
