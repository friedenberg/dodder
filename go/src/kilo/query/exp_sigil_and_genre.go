package query

import (
	"sort"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type expSigilAndGenre struct {
	ids.Sigil
	ids.Genre

	exp

	Hidden sku.Query
}

func (expSigilAndGenre *expSigilAndGenre) IsEmpty() bool {
	return expSigilAndGenre.Sigil == ids.SigilUnknown &&
		expSigilAndGenre.Genre.IsEmpty() &&
		expSigilAndGenre.exp.IsEmpty()
}

func (expSigilAndGenre *expSigilAndGenre) GetSigil() ids.Sigil {
	return expSigilAndGenre.Sigil
}

func (expSigilAndGenre *expSigilAndGenre) addPinnedObjectId(
	b *buildState,
	k pinnedObjectId,
) (err error) {
	if err = expSigilAndGenre.addExactObjectId(b, k.ObjectId, k.Sigil); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (expSigilAndGenre *expSigilAndGenre) addExactObjectId(
	b *buildState,
	k ObjectId,
	sigil ids.Sigil,
) (err error) {
	if k.ObjectId == nil {
		err = errors.ErrorWithStackf("nil object id")
		return
	}

	expSigilAndGenre.Sigil.Add(sigil)
	expSigilAndGenre.expObjectIds.internal[k.GetObjectId().String()] = k
	expSigilAndGenre.Genre.Add(genres.Must(k))

	return
}

func (expSigilAndGenre *expSigilAndGenre) ContainsObjectId(
	k *ids.ObjectId,
) bool {
	if !expSigilAndGenre.Genre.Contains(k.GetGenre()) {
		err := errors.ErrorWithStackf(
			"checking query %#v for object id %#v, %q, %q",
			expSigilAndGenre,
			k,
			expSigilAndGenre,
			k,
		)
		panic(err)
	}

	if len(expSigilAndGenre.expObjectIds.internal) == 0 {
		return false
	}

	_, ok := expSigilAndGenre.expObjectIds.internal[k.String()]

	return ok
}

func (a *expSigilAndGenre) Clone() (b *expSigilAndGenre) {
	b = &expSigilAndGenre{
		Sigil: a.Sigil,
		Genre: a.Genre,
		exp: exp{
			expObjectIds: expObjectIds{
				internal: make(
					map[string]ObjectId,
					len(a.expObjectIds.internal),
				),
				external: make(
					map[string]sku.ExternalObjectId,
					len(a.expObjectIds.external),
				),
			},
		},
		Hidden: a.Hidden,
	}

	bExp := a.expTagsOrTypes.Clone()
	b.expTagsOrTypes = *bExp

	for k, v := range a.expObjectIds.internal {
		b.expObjectIds.internal[k] = v
	}

	for k, v := range a.expObjectIds.external {
		b.expObjectIds.external[k] = v
	}

	return b
}

func (q *expSigilAndGenre) Add(m sku.Query) (err error) {
	q1, ok := m.(*expSigilAndGenre)

	if !ok {
		return q.expTagsOrTypes.Add(m)
	}

	if q1.Genre != q.Genre {
		err = errors.ErrorWithStackf(
			"expected %q but got %q",
			q.Genre,
			q1.Genre,
		)

		return
	}

	if err = q.Merge(q1); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (expSigilAndGenre *expSigilAndGenre) Merge(
	b *expSigilAndGenre,
) (err error) {
	expSigilAndGenre.Sigil.Add(b.Sigil)

	if expSigilAndGenre.expObjectIds.internal == nil {
		expSigilAndGenre.expObjectIds.internal = make(
			map[string]ObjectId,
			len(b.expObjectIds.internal),
		)
	}

	for _, k := range b.expObjectIds.internal {
		expSigilAndGenre.expObjectIds.internal[k.GetObjectId().String()] = k
	}

	if expSigilAndGenre.expObjectIds.external == nil {
		expSigilAndGenre.expObjectIds.external = make(
			map[string]sku.ExternalObjectId,
			len(b.expObjectIds.external),
		)
	}

	for _, k := range b.expObjectIds.external {
		expSigilAndGenre.expObjectIds.external[k.GetExternalObjectId().String()] = k
	}

	expSigilAndGenre.expTagsOrTypes.Children = append(
		expSigilAndGenre.expTagsOrTypes.Children,
		b.expTagsOrTypes.Children...)

	return
}

func (expSigilAndGenre *expSigilAndGenre) StringDebug() string {
	var sb strings.Builder

	if expSigilAndGenre.expObjectIds.internal == nil ||
		len(expSigilAndGenre.expObjectIds.internal) == 0 {
		sb.WriteString(expSigilAndGenre.expTagsOrTypes.StringDebug())
	} else {
		sb.WriteString("[[")

		first := true

		for _, k := range expSigilAndGenre.expObjectIds.internal {
			if !first {
				sb.WriteString(", ")
			}

			sb.WriteString(k.String())

			first = false
		}

		sb.WriteString(", ")
		sb.WriteString(expSigilAndGenre.expTagsOrTypes.StringDebug())
		sb.WriteString("]")
	}

	if expSigilAndGenre.IsEmpty() && !expSigilAndGenre.IsLatestOrUnknown() {
		sb.WriteString(expSigilAndGenre.Sigil.String())
	} else if !expSigilAndGenre.IsEmpty() {
		sb.WriteString(expSigilAndGenre.Sigil.String())
		sb.WriteString(expSigilAndGenre.Genre.String())
	}

	return sb.String()
}

func (expSigilAndGenre *expSigilAndGenre) SortedObjectIds() []string {
	out := make([]string, 0, expSigilAndGenre.expObjectIds.Len())

	for k := range expSigilAndGenre.expObjectIds.internal {
		out = append(out, k)
	}

	for k := range expSigilAndGenre.expObjectIds.external {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}

func (expSigilAndGenre *expSigilAndGenre) String() string {
	var sb strings.Builder

	e := expSigilAndGenre.expTagsOrTypes.String()

	oids := expSigilAndGenre.SortedObjectIds()

	if len(oids) == 0 {
		sb.WriteString(e)
	} else if len(oids) == 1 && e == "" {
		for _, k := range oids {
			sb.WriteString(k)
		}
	} else {
		sb.WriteString("[")

		first := true

		for _, k := range oids {
			if !first {
				sb.WriteString(", ")
			}

			sb.WriteString(k)

			first = false
		}

		if e != "" {
			sb.WriteString(", ")
			sb.WriteString(expSigilAndGenre.expTagsOrTypes.String())
		}

		sb.WriteString("]")
	}

	if expSigilAndGenre.Genre.IsEmpty() &&
		!expSigilAndGenre.IsLatestOrUnknown() {
		sb.WriteString(expSigilAndGenre.Sigil.String())
	} else if !expSigilAndGenre.Genre.IsEmpty() {
		sb.WriteString(expSigilAndGenre.Sigil.String())
		sb.WriteString(expSigilAndGenre.Genre.String())
	}

	return sb.String()
}

func (expSigilAndGenre *expSigilAndGenre) ShouldHide(
	objectGetter sku.TransactedGetter,
	objectIdString string,
) bool {
	_, ok := expSigilAndGenre.expObjectIds.internal[objectIdString]

	if expSigilAndGenre.IncludesHidden() || expSigilAndGenre.Hidden == nil ||
		ok {
		return false
	}

	return expSigilAndGenre.Hidden.ContainsSku(objectGetter)
}

func (expSigilAndGenre *expSigilAndGenre) ContainsSku(
	objectGetter sku.TransactedGetter,
) (ok bool) {
	object := objectGetter.GetSku()

	objectIdString := object.ObjectId.String()

	genre := genres.Must(object)

	if expSigilAndGenre.ShouldHide(object, objectIdString) {
		return
	}

	if !expSigilAndGenre.Genre.ContainsOneOf(genre) {
		return
	}

	if _, ok = expSigilAndGenre.expObjectIds.internal[objectIdString]; ok {
		return
	}

	if len(expSigilAndGenre.expTagsOrTypes.Children) == 0 {
		ok = len(expSigilAndGenre.expObjectIds.internal) == 0
		return
	} else if !expSigilAndGenre.expTagsOrTypes.ContainsSku(objectGetter) {
		return
	}

	ok = true

	return
}

func (expSigilAndGenre *expSigilAndGenre) ContainsExternalSku(
	el sku.ExternalLike,
) (ok bool) {
	sk := el.GetSku()

	g := genres.Must(sk)

	if !expSigilAndGenre.Genre.ContainsOneOf(g) {
		return
	}

	k := sk.ObjectId.String()

	if expSigilAndGenre.ShouldHide(el, k) {
		return
	}

	eoid := el.GetExternalObjectId().String()
	ui.Log().Print(
		eoid,
		expSigilAndGenre.expObjectIds.external,
		expSigilAndGenre.expObjectIds.internal,
	)

	if _, ok = expSigilAndGenre.expObjectIds.external[eoid]; ok {
		return
	}

	if _, ok = expSigilAndGenre.expObjectIds.external[k]; ok {
		return
	}

	if _, ok = expSigilAndGenre.expObjectIds.internal[k]; ok {
		return
	}

	if len(expSigilAndGenre.expTagsOrTypes.Children) == 0 {
		ok = expSigilAndGenre.expObjectIds.IsEmpty()
		return
	} else if !expSigilAndGenre.expTagsOrTypes.ContainsSku(el) {
		return
	}

	ok = true

	return
}
