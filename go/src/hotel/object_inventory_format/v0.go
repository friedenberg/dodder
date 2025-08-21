package object_inventory_format

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type v0 struct{}

func (f v0) FormatPersistentMetadata(
	w1 io.Writer,
	c FormatterContext,
	o Options,
) (n int64, err error) {
	m := c.GetMetadata()
	w := format.NewLineWriter()

	if o.Tai {
		w.WriteFormat("Tai %s", m.Tai)
	}

	w.WriteFormat("%s %s", genres.Blob, &m.BlobId)
	w.WriteFormat("%s %s", genres.Type, m.GetType())
	w.WriteFormat("Bezeichnung %s", m.Description)

	for _, e := range quiter.SortedValues(m.GetTags()) {
		w.WriteFormat("%s %s", genres.Tag, e)
	}

	if n, err = w.WriteTo(w1); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (f v0) ParsePersistentMetadata(
	r1 io.Reader,
	c ParserContext,
	_ Options,
) (n int64, err error) {
	m := c.GetMetadata()

	tags := ids.MakeTagMutableSet()

	r := bufio.NewReader(r1)

	typLineReader := ohio.MakeLineReaderIgnoreErrors(m.Type.Set)

	esa := quiter.MakeFuncSetString[ids.Tag, *ids.Tag](
		tags,
	)

	var g genres.Genre

	lr := format.MakeLineReaderConsumeEmpty(
		ohio.MakeLineReaderIterate(
			g.Set,
			ohio.MakeLineReaderKeyValues(
				map[string]interfaces.FuncSetString{
					"Tai":                m.Tai.Set,
					genres.Blob.String(): m.BlobId.Set,
					genres.Type.String(): typLineReader,
					"AkteTyp":            typLineReader,
					"Bezeichnung":        m.Description.Set,
					genres.Tag.String():  esa,
				},
			),
		),
	)

	if n, err = lr.ReadFrom(r); err != nil {
		err = errors.Wrap(err)
		return
	}

	m.SetTags(tags)

	return
}
