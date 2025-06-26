package query

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Query struct {
	sku.ExternalQueryOptions

	hidden           sku.Query
	optimizedQueries map[genres.Genre]*expSigilAndGenre
	userQueries      map[ids.Genre]*expSigilAndGenre
	types            ids.TypeMutableSet

	dotOperatorActive bool
	matchOnEmpty      bool

	defaultQuery *Query
}

func (q *Query) GetDefaultQuery() *Query {
	return q.defaultQuery
}

func (qg *Query) isDotOperatorActive() bool {
	if qg.dotOperatorActive {
		return true
	}

	for _, oq := range qg.optimizedQueries {
		if oq.Sigil.ContainsOneOf(ids.SigilExternal) {
			return true
		}
	}

	return false
}

type reducer interface {
	reduce(*buildState) error
}

func (qg *Query) reduce(b *buildState) (err error) {
	for _, q := range qg.userQueries {
		if err = q.reduce(b); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = qg.addOptimized(b, q); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	for _, q := range qg.optimizedQueries {
		if err = q.reduce(b); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (qg *Query) addExactExternalObjectId(
	b *buildState,
	k sku.ExternalObjectId,
) (err error) {
	if k == nil {
		err = errors.ErrorWithStackf("nil object id")
		return
	}

	q := b.makeQuery()

	q.Sigil.Add(ids.SigilExternal)
	q.Sigil.Add(ids.SigilLatest)
	q.Genre.Add(genres.Must(k))
	q.expObjectIds.external[k.String()] = k

	if err = qg.add(q); err != nil {
		err = errors.Wrap(err)
		return
	}

	qg.dotOperatorActive = true

	return
}

func (qg *Query) add(q *expSigilAndGenre) (err error) {
	existing, ok := qg.userQueries[q.Genre]

	if !ok {
		existing = &expSigilAndGenre{
			Hidden: qg.hidden,
			Genre:  q.Genre,
			exp: exp{
				expObjectIds: expObjectIds{
					internal: make(map[string]ObjectId),
				},
			},
		}
	}

	if err = existing.Add(q); err != nil {
		err = errors.Wrap(err)
		return
	}

	qg.userQueries[q.Genre] = existing

	return
}

func (qg *Query) addOptimized(b *buildState, q *expSigilAndGenre) (err error) {
	q = q.Clone()
	gs := q.Slice()

	if len(gs) == 0 {
		gs = b.defaultGenres.Slice()
	}

	for _, g := range gs {
		existing, ok := qg.optimizedQueries[g]

		if !ok {
			existing = b.makeQuery()
			existing.Genre = ids.MakeGenre(g)
		}

		if err = existing.Merge(q); err != nil {
			err = errors.Wrap(err)
			return
		}

		qg.optimizedQueries[g] = existing
	}

	return
}

func (qg *Query) isEmpty() bool {
	return len(qg.userQueries) == 0
}

func (queryGroup *Query) getExactlyOneExternalObjectId(
	permitInternal bool,
) (objectId ids.ObjectIdLike, sigil ids.Sigil, err error) {
	if len(queryGroup.optimizedQueries) != 1 {
		err = errors.ErrorWithStackf(
			"expected exactly 1 genre query but got %d",
			len(queryGroup.optimizedQueries),
		)

		return
	}

	var query *expSigilAndGenre

	for _, query = range queryGroup.optimizedQueries {
		break
	}

	if query.Sigil.ContainsOneOf(ids.SigilHistory) {
		err = errors.ErrorWithStackf(
			"sigil (%s) includes history, which may return multiple objects",
			query.Sigil,
		)

		return
	}

	oids := query.expObjectIds.internal
	oidsLen := len(oids)

	eoids := query.expObjectIds.external
	eoidsLen := len(eoids)

	switch {
	case eoidsLen == 0 && oidsLen == 1 && permitInternal:
		for _, k1 := range oids {
			objectId = k1
		}

	case eoidsLen == 1 && oidsLen == 0:
		for _, k1 := range eoids {
			objectId = k1.GetExternalObjectId()
		}

		sigil.Add(ids.SigilExternal)

	default:
		err = errors.ErrorWithStackf(
			"expected to exactly 1 object id or 1 external object id but got %d object ids and %d external object ids. Permit internal: %t",
			oidsLen,
			eoidsLen,
			permitInternal,
		)

		return
	}

	sigil = query.GetSigil()

	return
}

func (queryGroup *Query) getExactlyOneObjectId() (objectId *ids.ObjectId, sigil ids.Sigil, err error) {
	if len(queryGroup.optimizedQueries) != 1 {
		err = errors.ErrorWithStackf(
			"expected exactly 1 genre query but got %d",
			len(queryGroup.optimizedQueries),
		)

		return
	}

	var query *expSigilAndGenre

	for _, query = range queryGroup.optimizedQueries {
		break
	}

	if query.Sigil.ContainsOneOf(ids.SigilHistory) {
		err = errors.ErrorWithStackf(
			"sigil (%s) includes history, which may return multiple objects",
			query.Sigil,
		)

		return
	}

	oids := query.expObjectIds.internal
	oidsLen := len(oids)

	eoids := query.expObjectIds.external
	eoidsLen := len(eoids)

	switch {
	case eoidsLen == 0 && oidsLen == 1:
		for _, k1 := range oids {
			objectId = k1.GetObjectId()
		}

	default:
		err = errors.ErrorWithStackf(
			"expected to exactly 1 object id or 1 external object id but got %d object ids and %d external object ids",
			oidsLen,
			eoidsLen,
		)

		return
	}

	sigil = query.GetSigil()

	return
}

func (qg *Query) sortedUserQueries() []*expSigilAndGenre {
	out := make([]*expSigilAndGenre, 0, len(qg.userQueries))

	for _, g := range qg.userQueries {
		out = append(out, g)
	}

	sort.Slice(out, func(i, j int) bool {
		l, r := out[i].Genre, out[j].Genre

		if l.IsEmpty() {
			return false
		}

		if r.IsEmpty() {
			return true
		}

		return l < r
	})

	return out
}

func (qg *Query) containsSku(tg sku.TransactedGetter) (ok bool) {
	if qg.defaultQuery != nil &&
		!qg.defaultQuery.containsSku(tg) {
		return
	}

	sk := tg.GetSku()

	if len(qg.optimizedQueries) == 0 && qg.matchOnEmpty {
		ok = true
		return
	}

	g := sk.GetGenre()

	q, ok := qg.optimizedQueries[genres.Must(g)]

	if !ok || !q.ContainsSku(tg) {
		ok = false
		return
	}

	ok = true

	return
}
