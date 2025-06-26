package tag_paths

import (
	"fmt"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type Tags struct {
	Paths PathsWithTypes // TODO implement
	All   TagsWithParentsAndTypes
}

func (a *Tags) String() string {
	return fmt.Sprintf("[Paths: %s, All: %s]", a.Paths, a.All)
}

func (a *Tags) Reset() {
	// TODO pool *Path's
	a.Paths.Reset()
	a.All.Reset()
}

// TODO improve performance
func (a *Tags) ResetWith(b *Tags) {
	a.Paths = slices.Grow(a.Paths, len(b.Paths))

	for _, p := range b.Paths {
		a.AddPath(p.Clone())
	}
	// a.Paths = a.Paths[:cap(a.Paths)]
	// nPaths := copy(a.Paths, b.Paths)

	// a.All = slices.Grow(a.All, len(b.All))
	// a.All = a.All[:cap(a.All)]
	// nAll := copy(a.All, b.All)
	// ui.Debug().Print(nPaths, nAll, a, b)
}

func (a *Tags) AddSuperFrom(
	b *Tags,
	prefix *Tag,
) (err error) {
	for _, ep := range b.Paths {
		ui.Log().Print("adding", prefix, ep)
		if prefix.ComparePartial(ep.First()) == 0 {
			continue
		}

		prefixPath := makePath(prefix)
		prefixPath.Add(ep.Path...)

		c := &PathWithType{
			Path: prefixPath,
			Type: TypeSuper,
		}

		if err = a.AddPath(c); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (es *Tags) AddTagOld(e ids.Tag) (err error) {
	return es.AddTag(catgut.MakeFromString(e.String()))
}

func (es *Tags) AddTag(e *Tag) (err error) {
	if e.IsEmpty() {
		return
	}

	p := MakePathWithType(e)

	if err = es.AddPathWithType(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (es *Tags) AddSelf(e *Tag) (err error) {
	if e.IsEmpty() {
		return
	}

	p := MakePathWithType(e)
	p.Type = TypeSelf

	if err = es.AddPathWithType(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (es *Tags) AddPathWithType(pwt *PathWithType) (err error) {
	_, alreadyExists := es.Paths.AddPath(pwt)

	if alreadyExists {
		return
	}

	for _, e := range pwt.Path {
		if err = es.All.Add(e, pwt); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (es *Tags) AddPath(p *PathWithType) (err error) {
	if err = es.AddPathWithType(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Tags) Set(v string) (err error) {
	vs := strings.Split(v, ",")

	for _, v := range vs {
		var e ids.Tag

		if err = e.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

		es := catgut.MakeFromString(e.String())

		if err = s.AddTag(es); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
