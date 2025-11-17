package object_metadata

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

func StringSansTai(o *metadata) (str string) {
	sb := &strings.Builder{}

	sb.WriteString(" ")
	sb.WriteString(o.DigBlob.String())

	metadata := o.GetMetadata()

	t := metadata.GetType()

	if !t.IsEmpty() {
		sb.WriteString(" ")
		sb.WriteString(ids.FormattedString(metadata.GetType()))
	}

	es := metadata.GetTags()

	if es.Len() > 0 {
		sb.WriteString(" ")
		sb.WriteString(
			quiter.StringDelimiterSeparated[ids.Tag](
				" ",
				metadata.GetTags(),
			),
		)
	}

	b := metadata.GetDescription()

	if !b.IsEmpty() {
		sb.WriteString(" ")
		sb.WriteString("\"" + b.String() + "\"")
	}

	for field := range metadata.GetFields() {
		sb.WriteString(" ")
		fmt.Fprintf(sb, "%q=%q", field.Key, field.Value)
	}

	return sb.String()
}
