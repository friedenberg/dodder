package queries

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type expTagsOrTypes struct {
	Or       bool
	Negated  bool
	Exact    bool
	Hidden   bool
	Debug    bool
	Children []sku.Query
}

func (tagsOrTypes *expTagsOrTypes) Clone() (b *expTagsOrTypes) {
	b = &expTagsOrTypes{
		Or:      tagsOrTypes.Or,
		Negated: tagsOrTypes.Negated,
		Exact:   tagsOrTypes.Exact,
		Hidden:  tagsOrTypes.Hidden,
		Debug:   tagsOrTypes.Debug,
	}

	b.Children = make([]sku.Query, len(tagsOrTypes.Children))

	for i, c := range tagsOrTypes.Children {
		switch ct := c.(type) {
		case *expTagsOrTypes:
			b.Children[i] = ct.Clone()

		default:
			b.Children[i] = ct
		}
	}

	return b
}

func (tagsOrTypes *expTagsOrTypes) CollectTags(mes ids.TagSetMutable) {
	if tagsOrTypes.Or || tagsOrTypes.Negated {
		return
	}

	for _, m := range tagsOrTypes.Children {
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

func (tagsOrTypes *expTagsOrTypes) reduce(b *buildState) (err error) {
	if tagsOrTypes.Exact {
		for _, child := range tagsOrTypes.Children {
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

	chillen := make([]sku.Query, 0, len(tagsOrTypes.Children))

	for _, m := range tagsOrTypes.Children {
		switch mt := m.(type) {
		case *expTagsOrTypes:
			if err = mt.reduce(b); err != nil {
				err = errors.Wrap(err)
				return err
			}

			if len(mt.Children) == 0 {
				continue
			}

			if mt.Or == tagsOrTypes.Or && mt.Negated == tagsOrTypes.Negated && mt.Exact == tagsOrTypes.Exact {
				chillen = append(chillen, mt.Children...)
				continue
			}

		case reducer:
			if err = mt.reduce(b); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		chillen = append(chillen, m)
	}

	tagsOrTypes.Children = chillen

	return err
}

func (tagsOrTypes *expTagsOrTypes) Add(m sku.Query) (err error) {
	switch mt := m.(type) {
	case *expTagsOrTypes:

	case *ObjectId:
		mt.Exact = tagsOrTypes.Exact
	}

	tagsOrTypes.Children = append(tagsOrTypes.Children, m)

	return err
}

func (tagsOrTypes *expTagsOrTypes) Operator() rune {
	if tagsOrTypes.Or {
		return doddish.OpOr
	} else {
		return doddish.OpAnd
	}
}

func (tagsOrTypes *expTagsOrTypes) StringDebug() string {
	var sb strings.Builder

	op := tagsOrTypes.Operator()

	if tagsOrTypes.Negated {
		sb.WriteRune('^')
	}

	sb.WriteRune(doddish.OpGroupOpen)
	fmt.Fprintf(&sb, "(%d)", len(tagsOrTypes.Children))

	for i, m := range tagsOrTypes.Children {
		if i > 0 {
			sb.WriteRune(op)
		}

		sb.WriteString(m.String())
	}

	sb.WriteRune(doddish.OpGroupClose)

	return sb.String()
}

func (tagsOrTypes *expTagsOrTypes) String() string {
	if tagsOrTypes.Hidden {
		return ""
	}

	l := len(tagsOrTypes.Children)

	if l == 0 {
		return ""
	}

	var sb strings.Builder

	op := tagsOrTypes.Operator()

	if tagsOrTypes.Negated {
		sb.WriteRune('^')
	}

	switch l {
	case 1:
		sb.WriteString(tagsOrTypes.Children[0].String())

	default:
		sb.WriteRune(doddish.OpGroupOpen)

		for i, m := range tagsOrTypes.Children {
			if i > 0 {
				sb.WriteRune(op)
			}

			sb.WriteString(m.String())
		}

		sb.WriteRune(doddish.OpGroupClose)
	}

	return sb.String()
}

func (tagsOrTypes *expTagsOrTypes) negateIfNecessary(value bool) bool {
	if tagsOrTypes.Negated {
		return !value
	} else {
		return value
	}
}

func (tagsOrTypes *expTagsOrTypes) ContainsSku(
	objectGetter sku.TransactedGetter,
) (ok bool) {
	if len(tagsOrTypes.Children) == 0 {
		ok = tagsOrTypes.negateIfNecessary(true)
		return ok
	}

	if tagsOrTypes.Or {
		ok = tagsOrTypes.containsMatchableOr(objectGetter)
	} else {
		ok = tagsOrTypes.containsMatchableAnd(objectGetter)
	}

	return ok
}

func (tagsOrTypes *expTagsOrTypes) containsMatchableAnd(
	tg sku.TransactedGetter,
) bool {
	for _, m := range tagsOrTypes.Children {
		if !m.ContainsSku(tg) {
			return tagsOrTypes.negateIfNecessary(false)
		}
	}

	return tagsOrTypes.negateIfNecessary(true)
}

func (tagsOrTypes *expTagsOrTypes) containsMatchableOr(
	tg sku.TransactedGetter,
) bool {
	for _, m := range tagsOrTypes.Children {
		if m.ContainsSku(tg) {
			return tagsOrTypes.negateIfNecessary(true)
		}
	}

	return tagsOrTypes.negateIfNecessary(false)
}

func (tagsOrTypes *expTagsOrTypes) Each(
	f interfaces.FuncIter[sku.Query],
) (err error) {
	for _, m := range tagsOrTypes.Children {
		if err = f(m); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
