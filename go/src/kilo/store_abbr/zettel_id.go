package store_abbr

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type indexZettelId struct {
	readFunc func() error
	Heads    interfaces.MutableTridex
	Tails    interfaces.MutableTridex
}

func (ih *indexZettelId) Add(h *ids.ZettelId) (err error) {
	ih.Heads.Add(h.GetHead())
	ih.Tails.Add(h.GetTail())
	return
}

func (ih *indexZettelId) Exists(parts [3]string) (err error) {
	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !ih.Heads.ContainsExpansion(parts[0]) {
		err = collections.MakeErrNotFoundString(parts[0])
		return
	}

	if !ih.Tails.ContainsExpansion(parts[2]) {
		err = collections.MakeErrNotFoundString(parts[2])
		return
	}

	return
}

func (ih *indexZettelId) ExpandStringString(in string) (out string, err error) {
	var h *ids.ZettelId

	if h, err = ih.ExpandString(in); err != nil {
		err = errors.Wrap(err)
		return
	}

	out = h.String()

	return
}

func (ih *indexZettelId) ExpandString(s string) (h *ids.ZettelId, err error) {
	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var ha *ids.ZettelId

	if ha, err = ids.MakeZettelId(s); err != nil {
		err = errors.Wrap(err)
		return
	}

	if h, err = ih.Expand(ha); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (ih *indexZettelId) Expand(
	hAbbr *ids.ZettelId,
) (h *ids.ZettelId, err error) {
	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	head := ih.Heads.Expand(hAbbr.GetHead())
	tail := ih.Tails.Expand(hAbbr.GetTail())

	if h, err = ids.MakeZettelIdFromHeadAndTail(head, tail); err != nil {
		err = errors.Wrapf(err, "{Abbreviated: '%s'}", hAbbr)
		return
	}

	return
}

func (ih *indexZettelId) Abbreviate(
	id ids.Abbreviatable,
) (v string, err error) {
	var h ids.ZettelId

	switch idt := id.(type) {
	case ids.ZettelId:
		h = idt

	case *ids.ObjectId:
		if idt.GetGenre() != genres.Zettel {
			err = genres.MakeErrUnsupportedGenre(idt)
			return
		}

		if err = h.Set(idt.String()); err != nil {
			err = errors.Wrap(err)
			return
		}

	default:
		err = errors.ErrorWithStackf("unsupported type %T: %q", idt, idt)
		return
	}

	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	head := ih.Heads.Abbreviate(h.GetHead())
	tail := ih.Tails.Abbreviate(h.GetTail())

	if head == "" {
		v = h.String()
		return
	}

	if tail == "" {
		v = h.String()
		return
	}

	v = fmt.Sprintf("%s/%s", head, tail)

	return
}
