package store_abbr

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// TODO refactor this whole thing to be simpler and get rid of generics
type AbbrStoreMutableGeneric[V any, VPtr interfaces.Ptr[V]] interface {
	Add(VPtr) error
}

type AbbrStoreCompleteGeneric[V any, VPtr interfaces.Ptr[V]] interface {
	sku.AbbrStoreGeneric[V, VPtr]
	AbbrStoreMutableGeneric[V, VPtr]
}

type indexNoAbbr[
	V interfaces.Stringer,
	VPtr interfaces.SetterPtr[V],
] struct {
	sku.AbbrStoreGeneric[V, VPtr]
}

func (ih indexNoAbbr[V, VPtr]) Abbreviate(h V) (v string, err error) {
	v = h.String()
	return
}

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

func (ih *indexZettelId) Abbreviate(id ids.Abbreviatable) (v string, err error) {
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

type indexNotZettelId[
	K any,
	KPtr interfaces.StringerSetterPtr[K],
] struct {
	readFunc  func() error
	ObjectIds interfaces.MutableTridex
}

func (ih *indexNotZettelId[K, KPtr]) Add(k KPtr) (err error) {
	ih.ObjectIds.Add(k.String())
	return
}

func (ih *indexNotZettelId[K, KPtr]) Exists(parts [3]string) (err error) {
	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !ih.ObjectIds.ContainsExpansion(parts[2]) {
		err = collections.MakeErrNotFoundString(parts[2])
		return
	}

	return
}

func (ih *indexNotZettelId[K, KPtr]) ExpandStringString(
	in string,
) (out string, err error) {
	var k KPtr

	if k, err = ih.ExpandString(in); err != nil {
		err = errors.Wrap(err)
		return
	}

	out = k.String()

	return
}

func (ih *indexNotZettelId[K, KPtr]) ExpandString(s string) (k KPtr, err error) {
	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var k1 K
	k = &k1

	if err = k.Set(s); err != nil {
		err = errors.Wrap(err)
		return
	}

	if k, err = ih.Expand(k); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (ih *indexNotZettelId[K, KPtr]) Expand(
	abbr KPtr,
) (exp KPtr, err error) {
	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	ex := ih.ObjectIds.Expand(abbr.String())

	if ex == "" {
		// TODO-P4 should try to use the expansion if possible
		ex = abbr.String()
	}

	var k K
	exp = &k

	if err = exp.Set(ex); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (ih *indexNotZettelId[K, KPtr]) Abbreviate(
	k ids.Abbreviatable,
) (v string, err error) {
	if err = ih.readFunc(); err != nil {
		err = errors.Wrap(err)
		return
	}

	v = ih.ObjectIds.Abbreviate(k.String())

	return
}
