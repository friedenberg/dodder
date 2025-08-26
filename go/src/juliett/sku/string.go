package sku

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func String(object *Transacted) (str string) {
	return StringMetadataTaiMerkle(object)
}

func StringTaiGenreObjectIdShaBlob(o *Transacted) (str string) {
	str = fmt.Sprintf(
		"%s %s %s %s %s",
		o.GetTai(),
		o.GetGenre(),
		o.GetObjectId(),
		o.GetObjectDigest(),
		o.GetBlobDigest(),
	)

	return
}

func StringObjectIdBlobMetadataSansTai(o *Transacted) (str string) {
	str = fmt.Sprintf(
		"%s %s %s",
		o.GetObjectId(),
		o.GetBlobDigest(),
		StringMetadataSansTai(o),
	)

	return
}

func StringMetadataTaiMerkle(object *Transacted) (str string) {
	tai := object.GetTai()
	taiFormatted := ids.MakeTaiRFC3339Value(tai)

	return fmt.Sprintf(
		"%s (%s) %s",
		tai,
		taiFormatted,
		StringMetadataSansTaiMerkle(object),
	)
}

func StringMetadataSansTai(object *Transacted) (str string) {
	sb := &strings.Builder{}

	sb.WriteString(object.GetGenre().GetGenreString())

	sb.WriteString(" ")
	sb.WriteString(object.GetObjectId().String())

	sb.WriteString(" ")
	sb.WriteString(object.GetExternalObjectId().String())

	sb.WriteString(" ")
	sb.WriteString(merkle_ids.Format(object.GetBlobDigest()))

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
			quiter.StringDelimiterSeparated(
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

// TODO switch to using fmt.Fprintf for panic recovery
func StringMetadataSansTaiMerkle(object *Transacted) (str string) {
	sb := &strings.Builder{}

	sb.WriteString(object.GetGenre().GetGenreString())

	sb.WriteString(" ")
	sb.WriteString(object.GetObjectId().String())

	sb.WriteString(" ")
	sb.WriteString(object.GetExternalObjectId().String())

	sb.WriteString(" ")
	fmt.Fprintf(sb, "%s", object.Metadata.GetRepoPubKey())

	sb.WriteString(" ")
	fmt.Fprintf(sb, "%s", object.Metadata.GetObjectSig())

	sb.WriteString(" ")
	fmt.Fprintf(sb, "%s", object.Metadata.GetObjectDigest())

	sb.WriteString(" ")
	sb.WriteString(merkle_ids.Format(object.GetBlobDigest()))

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
			quiter.StringDelimiterSeparated(
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

func StringMetadataSansTaiMerkle2(
	object *object_metadata.Metadata,
) (str string) {
	sb := &strings.Builder{}

	sb.WriteString(" ")
	fmt.Fprintf(sb, "%s", object.GetRepoPubKey())

	sb.WriteString(" ")
	fmt.Fprintf(sb, "%s", object.GetObjectSig())

	sb.WriteString(" ")
	fmt.Fprintf(sb, "%s", object.GetObjectDigest())

	sb.WriteString(" ")
	sb.WriteString(merkle_ids.Format(object.GetBlobDigest()))

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
			quiter.StringDelimiterSeparated(
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
