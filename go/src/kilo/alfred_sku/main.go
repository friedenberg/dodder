package alfred_sku

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/alfred"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Writer struct {
	alfredWriter alfred.Writer
	abbr         ids.Abbr
	organizeFmt  interfaces.StringEncoderTo[*sku.Transacted]
	alfred.ItemPool
}

func New(
	out io.Writer,
	abbr ids.Abbr,
	organizeFmt interfaces.StringEncoderTo[*sku.Transacted],
	aw alfred.Writer,
	itemPool alfred.ItemPool,
) (w *Writer, err error) {
	w = &Writer{
		abbr:         abbr,
		alfredWriter: aw,
		organizeFmt:  organizeFmt,
		ItemPool:     itemPool,
	}

	return
}

func (writer *Writer) SetWriter(alfredWriter alfred.Writer) {
	writer.alfredWriter = alfredWriter
}

func (writer *Writer) PrintOne(object *sku.Transacted) (err error) {
	var item *alfred.Item
	g := object.GetGenre()

	switch g {
	case genres.Zettel:
		item = writer.zettelToItem(object)

	case genres.Tag:
		var tag ids.Tag

		if err = tag.Set(object.ObjectId.String()); err != nil {
			err = errors.Wrap(err)
			return
		}

		item = writer.emitTag(object, &tag)

	default:
		item = writer.Get()
		item.Title = fmt.Sprintf("not implemented for genre: %q", g)
		item.Subtitle = sku.StringTaiGenreObjectIdShaBlob(object)
	}

	writer.alfredWriter.WriteItem(item)

	return
}

func (writer *Writer) WriteZettelId(e ids.ZettelId) (n int64, err error) {
	item := writer.zettelIdToItem(e)
	writer.alfredWriter.WriteItem(item)
	return
}

func (writer *Writer) WriteError(in error) (n int64, out error) {
	if in == nil {
		return 0, nil
	}

	var errorGroupBuilder errors.GroupBuilder

	if errors.As(in, &errorGroupBuilder) {
		for _, err := range errorGroupBuilder.Errors() {
			item := writer.errorToItem(err)
			writer.alfredWriter.WriteItem(item)
		}
	} else {
		item := writer.errorToItem(in)
		writer.alfredWriter.WriteItem(item)
	}

	return
}

func (writer Writer) Close() (err error) {
	return writer.alfredWriter.Close()
}

func (writer *Writer) addCommonMatches(
	object *sku.Transacted,
	item *alfred.Item,
) {
	k := &object.ObjectId
	ks := k.String()

	matchBuilder := alfred.GetPoolMatchBuilder().Get()
	defer alfred.GetPoolMatchBuilder().Put(matchBuilder)

	parts := k.PartsStrings()

	matchBuilder.AddMatches(ks)
	matchBuilder.AddMatchBytes(parts.Left.Bytes())
	matchBuilder.AddMatchBytes(parts.Right.Bytes())

	errors.PanicIfError(writer.abbr.AbbreviateZettelIdOnly(k))
	matchBuilder.AddMatches(k.String())
	parts = k.PartsStrings()
	matchBuilder.AddMatchBytes(parts.Left.Bytes())
	matchBuilder.AddMatchBytes(parts.Right.Bytes())

	matchBuilder.AddMatches(object.GetMetadata().Description.String())
	matchBuilder.AddMatches(object.GetType().String())
	for e := range object.Metadata.GetTags().All() {
		expansion.ExpanderAll.Expand(
			func(v string) (err error) {
				matchBuilder.AddMatches(v)
				return
			},
			e.String(),
		)
	}

	t := object.GetType()

	expansion.ExpanderAll.Expand(
		func(v string) (err error) {
			matchBuilder.AddMatches(v)
			return
		},
		t.String(),
	)

	item.Match.Write(matchBuilder.Bytes())
	// a.Match.ReadFromBuffer(&mb.Buffer)
}

func (writer *Writer) zettelToItem(
	object *sku.Transacted,
) (item *alfred.Item) {
	item = writer.Get()

	item.Title = object.Metadata.Description.String()

	es := quiter.StringCommaSeparated(
		object.Metadata.GetTags(),
	)

	k := &object.ObjectId
	ks := k.String()

	if item.Title == "" {
		item.Title = ks
		item.Subtitle = es
	} else {
		item.Subtitle = fmt.Sprintf("%s: %s %s", object.Metadata.Type, ks, es)
	}

	item.Arg = ks

	writer.addCommonMatches(object, item)

	item.Text.Copy = ks
	item.Uid = "dodder://" + ks

	{
		var sb strings.Builder

		if _, err := writer.organizeFmt.EncodeStringTo(object, &sb); err != nil {
			item = writer.errorToItem(err)
			return
		}

		item.Mods["alt"] = alfred.Mod{
			Valid:    true,
			Arg:      sb.String(),
			Subtitle: sb.String(),
		}
	}

	return
}

func (writer *Writer) emitTag(
	object *sku.Transacted,
	tag *ids.Tag,
) (item *alfred.Item) {
	item = writer.Get()

	item.Title = "@" + tag.String()

	item.Arg = tag.String()

	writer.addCommonMatches(object, item)

	item.Text.Copy = tag.String()
	item.Uid = "dodder://" + tag.String()

	return
}

func (writer *Writer) errorToItem(err error) (a *alfred.Item) {
	a = writer.Get()

	a.Title = errors.Unwrap(err).Error()

	return
}

func (writer *Writer) zettelIdToItem(e ids.ZettelId) (a *alfred.Item) {
	a = writer.Get()

	a.Title = e.String()

	a.Arg = e.String()

	mb := alfred.GetPoolMatchBuilder().Get()
	defer alfred.GetPoolMatchBuilder().Put(mb)

	mb.AddMatch(e.String())
	mb.AddMatch(e.GetHead())
	mb.AddMatch(e.GetTail())

	a.Match.ReadFromBuffer(&mb.Buffer)

	a.Text.Copy = e.String()
	a.Uid = "dodder://" + e.String()

	return
}
