package store_abbr

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type indexNotZettelId[
	ID any,
	ID_PTR interfaces.StringerSetterPtr[ID],
] struct {
	readFunc  func() error
	ObjectIds interfaces.MutableTridex
}

func (index *indexNotZettelId[ID, ID_PTR]) Add(k ID_PTR) (err error) {
	index.ObjectIds.Add(k.String())
	return err
}

func (index *indexNotZettelId[ID, ID_PTR]) Exists(parts [3]string) (err error) {
	if err = index.readFunc(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if !index.ObjectIds.ContainsExpansion(parts[2]) {
		err = collections.MakeErrNotFoundString(parts[2])
		return err
	}

	return err
}

func (index *indexNotZettelId[ID, ID_PTR]) ExpandStringString(
	in string,
) (out string, err error) {
	var k ID_PTR

	if k, err = index.ExpandString(in); err != nil {
		err = errors.Wrap(err)
		return out, err
	}

	out = k.String()

	return out, err
}

func (index *indexNotZettelId[ID, ID_PTR]) ExpandString(
	value string,
) (id ID_PTR, err error) {
	if err = index.readFunc(); err != nil {
		err = errors.Wrap(err)
		return id, err
	}

	var k1 ID
	id = &k1

	if err = id.Set(value); err != nil {
		err = errors.Wrap(err)
		return id, err
	}

	if id, err = index.Expand(id); err != nil {
		err = errors.Wrap(err)
		return id, err
	}

	return id, err
}

func (index *indexNotZettelId[ID, ID_PTR]) Expand(
	abbr ID_PTR,
) (exp ID_PTR, err error) {
	if err = index.readFunc(); err != nil {
		err = errors.Wrap(err)
		return exp, err
	}

	ex := index.ObjectIds.Expand(abbr.String())

	if ex == "" {
		// TODO-P4 should try to use the expansion if possible
		ex = abbr.String()
	}

	var k ID
	exp = &k

	if err = exp.Set(ex); err != nil {
		err = errors.Wrap(err)
		return exp, err
	}

	return exp, err
}

func (index *indexNotZettelId[ID, ID_PTR]) Abbreviate(
	k ids.Abbreviatable,
) (v string, err error) {
	if err = index.readFunc(); err != nil {
		err = errors.Wrap(err)
		return v, err
	}

	v = index.ObjectIds.Abbreviate(k.String())

	return v, err
}
