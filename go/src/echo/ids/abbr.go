package ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

type (
	// TODO use catgut.String
	FuncExpandString     func(string) (string, error)
	FuncAbbreviateString func(Abbreviatable) (string, error)

	Abbr struct {
		Sha      abbrOne
		ZettelId abbrOne
	}

	abbrOne struct {
		Expand     FuncExpandString
		Abbreviate FuncAbbreviateString
	}
)

func DontExpandString(v string) (string, error) {
	return v, nil
}

func DontAbbreviateString[VPtr interfaces.Stringer](k VPtr) (string, error) {
	return k.String(), nil
}

func (a Abbr) ExpanderFor(g genres.Genre) FuncExpandString {
	switch g {
	case genres.Zettel:
		return a.ZettelId.Expand

		// TODO add repo abbreviation
	case genres.Tag, genres.Type, genres.Repo:
		return DontExpandString

	default:
		return nil
	}
}

func (a Abbr) LenHeadAndTail(
	in *ObjectId,
) (head, tail int, err error) {
	if in.GetGenre() != genres.Zettel || a.ZettelId.Abbreviate == nil {
		head, tail = in.LenHeadAndTail()
		return
	}

	var h ZettelId

	if err = h.Set(in.String()); err != nil {
		err = nil
		return
	}

	var abbr string

	if abbr, err = a.ZettelId.Abbreviate(h); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = h.Set(abbr); err != nil {
		err = errors.Wrap(err)
		return
	}

	head = len(h.GetHead())
	tail = len(h.GetTail())

	return
}

func (a Abbr) AbbreviateZettelIdOnly(
	in *ObjectId,
) (err error) {
	if in.GetGenre() != genres.Zettel || in.IsVirtual() {
		return
	}

	var getAbbr FuncAbbreviateString

	var h ZettelId

	if err = h.Set(in.String()); err != nil {
		err = nil
		return
	}

	getAbbr = a.ZettelId.Abbreviate

	var abbr string

	if abbr, err = getAbbr(h); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = in.SetWithGenre(abbr, h); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a Abbr) ExpandZettelIdOnly(
	in *ObjectId,
) (err error) {
	if in.GetGenre() != genres.Zettel || a.ZettelId.Expand == nil {
		return
	}

	var h ZettelId

	if err = h.Set(in.String()); err != nil {
		err = nil
		return
	}

	var ex string

	if ex, err = a.ZettelId.Expand(h.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = in.SetWithGenre(ex, h); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a Abbr) AbbreviateObjectId(
	in *ObjectId,
	out *ObjectId,
) (err error) {
	var getAbbr FuncAbbreviateString

	switch in.GetGenre() {
	case genres.Zettel:
		getAbbr = a.ZettelId.Abbreviate

	case genres.Tag, genres.Type, genres.Repo:
		getAbbr = DontAbbreviateString

	case genres.Config:
		out.ResetWith(in)
		return

	default:
		err = errors.ErrorWithStackf("unsupported object id: %q, %T", in, in)
		return
	}

	var abbr string

	if abbr, err = getAbbr(in); err != nil {
		err = nil
		out.ResetWith(in)
		// err = errors.Wrap(err)
		return
	}

	if err = out.SetWithGenre(abbr, in); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
