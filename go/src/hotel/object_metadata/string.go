package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

func StringSansTai(metadata *metadata) string {
	stringBuilder := &strings.Builder{}

	stringBuilder.WriteString(" ")
	stringBuilder.WriteString(metadata.DigBlob.String())

	tipe := metadata.GetType()

	if !tipe.IsEmpty() {
		stringBuilder.WriteString(" ")
		stringBuilder.WriteString(ids.FormattedString(metadata.GetType()))
	}

	tags := metadata.GetTags()

	if tags.Len() > 0 {
		stringBuilder.WriteString(" ")
		stringBuilder.WriteString(
			quiter.StringDelimiterSeparated[ids.Tag](
				" ",
				metadata.GetTags(),
			),
		)
	}

	description := metadata.GetDescription()

	if !description.IsEmpty() {
		stringBuilder.WriteString(" ")
		stringBuilder.WriteString("\"" + description.String() + "\"")
	}

	for field := range metadata.GetIndex().GetFields() {
		stringBuilder.WriteString(" ")
		fmt.Fprintf(stringBuilder, "%q=%q", field.Key, field.Value)
	}

	return stringBuilder.String()
}
