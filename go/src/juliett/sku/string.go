package sku

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func String(object *Transacted) (str string) {
	return StringMetadataTaiMerkle(object)
}

func StringTaiGenreObjectIdObjectDigestBlobDigest(
	object *Transacted,
) (str string) {
	if object == nil {
		return "nil object!"
	}

	str = fmt.Sprintf(
		"%s %s %s %s %s",
		object.GetTai(),
		object.GetGenre(),
		object.GetObjectId(),
		object.GetObjectDigest(),
		object.GetBlobDigest(),
	)

	return
}

func StringObjectIdBlobMetadataSansTai(object *Transacted) (str string) {
	if object == nil {
		return "nil object!"
	}

	str = fmt.Sprintf(
		"%s %s %s",
		object.GetObjectId(),
		object.GetBlobDigest(),
		StringMetadataSansTai(object),
	)

	return
}

func StringMetadataTaiMerkle(object *Transacted) (str string) {
	if object == nil {
		return "nil object!"
	}

	tai := object.GetTai()
	taiFormatted := ids.MakeTaiRFC3339Value(tai)

	return fmt.Sprintf(
		"%s (%s) %s",
		tai,
		taiFormatted,
		StringMetadataSansTaiMerkle(object),
	)
}

func writeMarklIdWithFormatIfNecessary(
	stringBuilder *strings.Builder,
	id interfaces.MarklId,
) {
	if id.IsNull() {
		return
	}

	stringBuilder.WriteString(" ")
	stringBuilder.WriteString(id.StringWithFormat())
}

func StringMetadataSansTai(object *Transacted) (str string) {
	sb := &strings.Builder{}

	sb.WriteString(object.GetGenre().GetGenreString())

	sb.WriteString(" ")
	sb.WriteString(object.GetObjectId().String())

	sb.WriteString(" ")
	sb.WriteString(object.GetExternalObjectId().String())

	writeMarklIdWithFormatIfNecessary(sb, object.GetBlobDigest())

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

	writeMarklIdWithFormatIfNecessary(sb, object.Metadata.GetRepoPubKey())
	writeMarklIdWithFormatIfNecessary(sb, object.Metadata.GetObjectSig())
	writeMarklIdWithFormatIfNecessary(sb, object.Metadata.GetMotherObjectSig())
	writeMarklIdWithFormatIfNecessary(sb, object.GetObjectDigest())
	writeMarklIdWithFormatIfNecessary(sb, object.GetBlobDigest())

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

	writeMarklIdWithFormatIfNecessary(sb, object.GetRepoPubKey())
	writeMarklIdWithFormatIfNecessary(sb, object.GetObjectSig())
	writeMarklIdWithFormatIfNecessary(sb, object.GetObjectDigest())
	writeMarklIdWithFormatIfNecessary(sb, object.GetBlobDigest())

	metadata := object.GetMetadata()

	tipe := metadata.GetType()

	if !tipe.IsEmpty() {
		sb.WriteString(" ")
		sb.WriteString(ids.FormattedString(metadata.GetType()))
	}

	es := metadata.GetTags()

	if es.Len() > 0 {
		sb.WriteString(" ")
		sb.WriteString(
			quiter.StringDelimiterSeparated(
				" ",
				metadata.GetTags(),
			),
		)
	}

	b := metadata.Description

	if !b.IsEmpty() {
		sb.WriteString(" ")
		sb.WriteString("\"" + b.String() + "\"")
	}

	for _, field := range metadata.Fields {
		sb.WriteString(" ")
		fmt.Fprintf(sb, "%q=%q", field.Key, field.Value)
	}

	return sb.String()
}
