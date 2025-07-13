package organize_text

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type PrefixSet struct {
	count    int
	innerMap map[string]objSet
}

type Segments struct {
	Ungrouped objSet
	Grouped   PrefixSet
}

func MakePrefixSet(c int) (s PrefixSet) {
	s.innerMap = make(map[string]objSet, c)
	return s
}

func MakePrefixSetFrom(
	ts objSet,
) (s PrefixSet) {
	s = MakePrefixSet(ts.Len())
	ts.Each(s.Add)
	return
}

func (s PrefixSet) Len() int {
	return s.count
}

func (s *PrefixSet) AddSku(object sku.SkuType) (err error) {
	if object.GetState() == checked_out_state.Unknown {
		err = errors.ErrorWithStackf(
			"unacceptable state: %s",
			object.GetState(),
		)
		err = errors.Wrapf(err, "Sku: %s", sku.String(object.GetSku()))
		return
	}

	o := obj{
		sku: sku.CloneSkuType(object),
	}

	if err = s.Add(&o); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// this splits on right-expanded
func (s *PrefixSet) Add(z *obj) (err error) {
	es := ids.Expanded(
		z.GetSkuExternal().GetMetadata().Cache.GetImplicitTags(),
		expansion.ExpanderRight,
	).CloneMutableSetPtrLike()

	if err = z.GetSkuExternal().GetMetadata().Cache.GetExpandedTags().EachPtr(
		es.AddPtr,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if es.Len() == 0 {
		s.addPair("", z)
		return
	}

	if err = es.Each(
		func(e ids.Tag) (err error) {
			s.addPair(e.String(), z)
			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a PrefixSet) Subtract(
	b objSet,
) (c PrefixSet) {
	c = MakePrefixSet(len(a.innerMap))

	for e, aSet := range a.innerMap {
		aSet.Each(
			func(z *obj) (err error) {
				if b.Contains(z) {
					return
				}

				c.addPair(e, z)

				return
			},
		)
	}

	return
}

func (s *PrefixSet) addPair(
	e string,
	z *obj,
) {
	if e == z.GetSkuExternal().ObjectId.String() {
		e = ""
	}

	existingSet, ok := s.innerMap[e]

	if !ok {
		existingSet = makeObjSet()
		s.innerMap[e] = existingSet
	}

	var existingObj *obj
	existingObj, ok = existingSet.Get(existingSet.Key(z))

	if ok && existingObj.tipe.IsDirectOrSelf() {
		z.tipe.SetDirect()
	} else if !ok {
		s.count += 1
	}

	existingSet.Add(z)
}

func (a PrefixSet) Each(
	f func(ids.Tag, objSet) error,
) (err error) {
	for e, ssz := range a.innerMap {
		var e1 ids.Tag

		if e != "" {
			e1 = ids.MustTag(e)
		}

		if err = f(e1, ssz); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return
		}
	}

	return
}

func (a PrefixSet) AllObjectSets() interfaces.Seq2[string, objSet] {
	return func(yield func(string, objSet) bool) {
		for i, os := range a.innerMap {
			if !yield(i, os) {
				break
			}
		}
	}
}

func (a PrefixSet) AllObjects() interfaces.Seq2[string, *obj] {
	return func(yield func(string, *obj) bool) {
		for i, os := range a.innerMap {
			for o := range os.All() {
				if !yield(i, o) {
					break
				}
			}
		}
	}
}

func (a PrefixSet) EachObject(
	f interfaces.FuncIter[*obj],
) error {
	return a.Each(
		func(_ ids.Tag, st objSet) (err error) {
			err = st.Each(
				func(z *obj) (err error) {
					err = f(z)
					return
				},
			)

			return
		},
	)
}

func (a PrefixSet) Match(
	e ids.Tag,
) (out Segments) {
	out.Ungrouped = makeObjSet()
	out.Grouped = MakePrefixSet(len(a.innerMap))

	for e1, zSet := range a.innerMap {
		if e1 == "" {
			continue
		}

		zSet.Each(
			func(z *obj) (err error) {
				es := z.GetSkuExternal().GetTags()

				intersection := ids.IntersectPrefixes(
					es,
					e,
				)

				exactMatch := intersection.Len() == 1 &&
					intersection.Any().Equals(e)

				if intersection.Len() == 0 && !exactMatch {
					return
				}

				for _, e2 := range quiter.Elements(intersection) {
					out.Grouped.addPair(e2.String(), z)
				}

				return
			},
		)
	}

	return
}

func (a PrefixSet) Subset(
	e ids.Tag,
) (out Segments) {
	out.Ungrouped = makeObjSet()
	out.Grouped = MakePrefixSet(len(a.innerMap))

	e2 := catgut.MakeFromString(e.String())

	for e1, zSet := range a.innerMap {
		if e1 == "" {
			continue
		}

		zSet.Each(
			func(z *obj) (err error) {
				ui.Log().Print(e2, z)
				intersection := z.GetSkuExternal().Metadata.Cache.TagPaths.All.GetMatching(
					e2,
				)
				hasDirect := false || len(intersection) == 0
				type match struct {
					string
					tag_paths.Type
				}
				toAddGrouped := make([]match, 0)

			OUTER:
				for _, e2Match := range intersection {
					e2s := e2Match.Tag.String()
					ui.Log().Print(e2Match.Tag)
					for _, e3 := range e2Match.Parents {
						toAddGrouped = append(toAddGrouped, match{
							string: e2s,
							Type:   e3.Type,
						})

						ui.Log().Print(e3)
						if e3.Type == tag_paths.TypeDirect &&
							e2Match.Tag.Len() == e2.Len() {
							hasDirect = true
							break OUTER
						}
					}
				}

				if hasDirect {
					out.Ungrouped.Add(z)
				} else {
					for _, e3 := range toAddGrouped {
						c := z.cloneWithType(e3.Type)
						out.Grouped.addPair(e3.string, c)
					}
				}

				return
			},
		)
	}

	return
}
