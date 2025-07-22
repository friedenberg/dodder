package sku_fmt

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func StringMetadataSansTai(object *sku.Transacted) (str string) {
	sb := &strings.Builder{}

	sb.WriteString(object.GetGenre().GetGenreString())

	sb.WriteString(" ")
	sb.WriteString(object.GetObjectId().String())

	sb.WriteString(" ")
	sb.WriteString(digests.Format(object.GetBlobSha()))

	m := object.GetMetadata()

	t := m.GetType()

	if !t.IsEmpty() {
		sb.WriteString(" ")
		sb.WriteString(ids.FormattedString(m.GetType()))
	}

	es := m.GetTags()

	if es.Len() > 0 {
		sb.WriteString(" ")
		sb.WriteString(
			quiter.StringDelimiterSeparated[ids.Tag](
				" ",
				m.GetTags(),
			),
		)
	}

	b := m.Description

	if !b.IsEmpty() {
		sb.WriteString(" ")
		sb.WriteString("\"" + b.String() + "\"")
	}

	return sb.String()
}
