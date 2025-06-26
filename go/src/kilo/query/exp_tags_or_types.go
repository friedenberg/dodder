package query

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/box"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type expTagsOrTypes struct {
	Or       bool
	Negated  bool
	Exact    bool
	Hidden   bool
	Debug    bool
	Children []sku.Query
}

func (a *expTagsOrTypes) Clone() (b *expTagsOrTypes) {
	b = &expTagsOrTypes{
		Or:      a.Or,
		Negated: a.Negated,
		Exact:   a.Exact,
		Hidden:  a.Hidden,
		Debug:   a.Debug,
	}

	b.Children = make([]sku.Query, len(a.Children))

	for i, c := range a.Children {
		switch ct := c.(type) {
		case *expTagsOrTypes:
			b.Children[i] = ct.Clone()

		default:
			b.Children[i] = ct
		}
	}

	return b
}

func (e *expTagsOrTypes) CollectTags(mes ids.TagMutableSet) {
	if e.Or || e.Negated {
		return
	}

	for _, m := range e.Children {
		switch mt := m.(type) {
		case *expTagsOrTypes:
			mt.CollectTags(mes)

		case *ObjectId:
			if mt.GetGenre() != genres.Tag {
				continue
			}

			e := ids.MustTag(mt.GetObjectId().String())
			mes.Add(e)
		}
	}
}

func (e *expTagsOrTypes) reduce(b *buildState) (err error) {
	if e.Exact {
		for _, child := range e.Children {
			switch k := child.(type) {
			case *ObjectId:
				k.Exact = true

			case *expTagsOrTypes:
				k.Exact = true

			default:
				continue
			}
		}
	}

	chillen := make([]sku.Query, 0, len(e.Children))

	for _, m := range e.Children {
		switch mt := m.(type) {
		case *expTagsOrTypes:
			if err = mt.reduce(b); err != nil {
				err = errors.Wrap(err)
				return
			}

			if len(mt.Children) == 0 {
				continue
			}

			if mt.Or == e.Or && mt.Negated == e.Negated && mt.Exact == e.Exact {
				chillen = append(chillen, mt.Children...)
				continue
			}

		case reducer:
			if err = mt.reduce(b); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		chillen = append(chillen, m)
	}

	e.Children = chillen

	return
}

func (e *expTagsOrTypes) Add(m sku.Query) (err error) {
	switch mt := m.(type) {
	case *expTagsOrTypes:

	case *ObjectId:
		mt.Exact = e.Exact
	}

	e.Children = append(e.Children, m)

	return
}

func (matcher *expTagsOrTypes) Operator() rune {
	if matcher.Or {
		return box.OpOr
	} else {
		return box.OpAnd
	}
}

func (e *expTagsOrTypes) StringDebug() string {
	var sb strings.Builder

	op := e.Operator()

	if e.Negated {
		sb.WriteRune('^')
	}

	sb.WriteRune(box.OpGroupOpen)
	fmt.Fprintf(&sb, "(%d)", len(e.Children))

	for i, m := range e.Children {
		if i > 0 {
			sb.WriteRune(op)
		}

		sb.WriteString(m.String())
	}

	sb.WriteRune(box.OpGroupClose)

	return sb.String()
}

func (e *expTagsOrTypes) String() string {
	if e.Hidden {
		return ""
	}

	l := len(e.Children)

	if l == 0 {
		return ""
	}

	var sb strings.Builder

	op := e.Operator()

	if e.Negated {
		sb.WriteRune('^')
	}

	switch l {
	case 1:
		sb.WriteString(e.Children[0].String())

	default:
		sb.WriteRune(box.OpGroupOpen)

		for i, m := range e.Children {
			if i > 0 {
				sb.WriteRune(op)
			}

			sb.WriteString(m.String())
		}

		sb.WriteRune(box.OpGroupClose)
	}

	return sb.String()
}

func (m *expTagsOrTypes) negateIfNecessary(v bool) bool {
	if m.Negated {
		return !v
	} else {
		return v
	}
}

func (e *expTagsOrTypes) ContainsSku(tg sku.TransactedGetter) (ok bool) {
	if len(e.Children) == 0 {
		ok = e.negateIfNecessary(true)
		return
	}

	if e.Or {
		ok = e.containsMatchableOr(tg)
	} else {
		ok = e.containsMatchableAnd(tg)
	}

	return
}

func (e *expTagsOrTypes) containsMatchableAnd(tg sku.TransactedGetter) bool {
	for _, m := range e.Children {
		if !m.ContainsSku(tg) {
			return e.negateIfNecessary(false)
		}
	}

	return e.negateIfNecessary(true)
}

func (e *expTagsOrTypes) containsMatchableOr(tg sku.TransactedGetter) bool {
	for _, m := range e.Children {
		if m.ContainsSku(tg) {
			return e.negateIfNecessary(true)
		}
	}

	return e.negateIfNecessary(false)
}

func (e *expTagsOrTypes) Each(
	f interfaces.FuncIter[sku.Query],
) (err error) {
	for _, m := range e.Children {
		if err = f(m); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
