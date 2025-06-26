package alfred_sku

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/alfred"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (w *Writer) addCommonMatches(
	z *sku.Transacted,
	a *alfred.Item,
) {
	k := &z.ObjectId
	ks := k.String()

	mb := alfred.GetPoolMatchBuilder().Get()
	defer alfred.GetPoolMatchBuilder().Put(mb)

	parts := k.PartsStrings()

	mb.AddMatches(ks)
	mb.AddMatchBytes(parts.Left.Bytes())
	mb.AddMatchBytes(parts.Right.Bytes())

	errors.PanicIfError(w.abbr.AbbreviateZettelIdOnly(k))
	mb.AddMatches(k.String())
	parts = k.PartsStrings()
	mb.AddMatchBytes(parts.Left.Bytes())
	mb.AddMatchBytes(parts.Right.Bytes())

	mb.AddMatches(z.GetMetadata().Description.String())
	mb.AddMatches(z.GetType().String())
	z.Metadata.GetTags().Each(
		func(e ids.Tag) (err error) {
			expansion.ExpanderAll.Expand(
				func(v string) (err error) {
					mb.AddMatches(v)
					return
				},
				e.String(),
			)

			return
		},
	)

	t := z.GetType()

	expansion.ExpanderAll.Expand(
		func(v string) (err error) {
			mb.AddMatches(v)
			return
		},
		t.String(),
	)

	a.Match.Write(mb.Bytes())
	// a.Match.ReadFromBuffer(&mb.Buffer)
}

func (w *Writer) zettelToItem(
	z *sku.Transacted,
) (a *alfred.Item) {
	a = w.Get()

	a.Title = z.Metadata.Description.String()

	es := quiter.StringCommaSeparated(
		z.Metadata.GetTags(),
	)

	k := &z.ObjectId
	ks := k.String()

	if a.Title == "" {
		a.Title = ks
		a.Subtitle = es
	} else {
		a.Subtitle = fmt.Sprintf("%s: %s %s", z.Metadata.Type, ks, es)
	}

	a.Arg = ks

	w.addCommonMatches(z, a)

	a.Text.Copy = ks
	a.Uid = "dodder://" + ks

	{
		var sb strings.Builder

		if _, err := w.organizeFmt.EncodeStringTo(z, &sb); err != nil {
			a = w.errorToItem(err)
			return
		}

		a.Mods["alt"] = alfred.Mod{
			Valid:    true,
			Arg:      sb.String(),
			Subtitle: sb.String(),
		}
	}

	return
}

func (w *Writer) etikettToItem(
	z *sku.Transacted,
	e *ids.Tag,
) (a *alfred.Item) {
	a = w.Get()

	a.Title = "@" + e.String()

	a.Arg = e.String()

	w.addCommonMatches(z, a)

	a.Text.Copy = e.String()
	a.Uid = "dodder://" + e.String()

	return
}

func (w *Writer) errorToItem(err error) (a *alfred.Item) {
	a = w.Get()

	a.Title = errors.Unwrap(err).Error()

	return
}

func (w *Writer) zettelIdToItem(e ids.ZettelId) (a *alfred.Item) {
	a = w.Get()

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
