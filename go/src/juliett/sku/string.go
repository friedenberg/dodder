package sku

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

func String(o *Transacted) (str string) {
	return StringMetadataTai(o)
}

func StringTaiGenreObjectIdShaBlob(o *Transacted) (str string) {
	str = fmt.Sprintf(
		"%s %s %s %s %s",
		o.GetTai(),
		o.GetGenre(),
		o.GetObjectId(),
		o.GetObjectSha(),
		o.GetBlobSha(),
	)

	return
}

func StringObjectIdBlobMetadataSansTai(o *Transacted) (str string) {
	str = fmt.Sprintf(
		"%s %s %s",
		o.GetObjectId(),
		o.GetBlobSha(),
		StringMetadataSansTai(o),
	)

	return
}

func StringMetadataTai(o *Transacted) (str string) {
	t := o.GetTai()
	t1 := ids.MakeTaiRFC3339Value(t)

	return fmt.Sprintf(
		"%s (%s) %s",
		t,
		t1,
		StringMetadataSansTai(o),
	)
}

func StringMetadataSansTai(o *Transacted) (str string) {
	sb := &strings.Builder{}

	sb.WriteString(o.GetGenre().GetGenreString())

	sb.WriteString(" ")
	sb.WriteString(o.GetObjectId().String())

	sb.WriteString(" ")
	sb.WriteString(o.GetExternalObjectId().String())

	sb.WriteString(" ")
	sb.WriteString(o.GetBlobSha().String())

	m := o.GetMetadata()

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

	for _, field := range m.Fields {
		sb.WriteString(" ")
		fmt.Fprintf(sb, "%q=%q", field.Key, field.Value)
	}

	return sb.String()
}
