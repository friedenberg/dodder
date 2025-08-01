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
	for element := range ts.All() {
		s.Add(element)
	}
	return
}

func (prefixSet PrefixSet) Len() int {
	return prefixSet.count
}

func (prefixSet *PrefixSet) AddSku(object sku.SkuType) (err error) {
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

	if err = prefixSet.Add(&o); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// this splits on right-expanded
func (prefixSet *PrefixSet) Add(z *obj) (err error) {
	es := ids.Expanded(
		z.GetSkuExternal().GetMetadata().Cache.GetImplicitTags(),
		expansion.ExpanderRight,
	).CloneMutableSetPtrLike()

	for e := range z.GetSkuExternal().GetMetadata().Cache.GetExpandedTags().AllPtr() {
		if err = es.AddPtr(e); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if es.Len() == 0 {
		prefixSet.addPair("", z)
		return
	}

	for e := range es.All() {
		prefixSet.addPair(e.String(), z)
	}

	return
}

func (prefixSet PrefixSet) Subtract(
	b objSet,
) (c PrefixSet) {
	c = MakePrefixSet(len(prefixSet.innerMap))

	for e, aSet := range prefixSet.innerMap {
		for z := range aSet.All() {
			if !b.Contains(z) {
				c.addPair(e, z)
			}
		}
	}

	return
}

func (prefixSet *PrefixSet) addPair(
	e string,
	z *obj,
) {
	if e == z.GetSkuExternal().ObjectId.String() {
		e = ""
	}

	existingSet, ok := prefixSet.innerMap[e]

	if !ok {
		existingSet = makeObjSet()
		prefixSet.innerMap[e] = existingSet
	}

	var existingObj *obj
	existingObj, ok = existingSet.Get(existingSet.Key(z))

	if ok && existingObj.tipe.IsDirectOrSelf() {
		z.tipe.SetDirect()
	} else if !ok {
		prefixSet.count += 1
	}

	existingSet.Add(z)
}

func (prefixSet PrefixSet) AllObjectSets() interfaces.Seq2[string, objSet] {
	return func(yield func(string, objSet) bool) {
		for tagString, objects := range prefixSet.innerMap {
			if !yield(tagString, objects) {
				break
			}
		}
	}
}

func (prefixSet PrefixSet) AllObjects() interfaces.Seq2[string, *obj] {
	return func(yield func(string, *obj) bool) {
		for tagString, objects := range prefixSet.innerMap {
			for object := range objects.All() {
				if !yield(tagString, object) {
					break
				}
			}
		}
	}
}

func (prefixSet PrefixSet) Match(
	e ids.Tag,
) (out Segments) {
	out.Ungrouped = makeObjSet()
	out.Grouped = MakePrefixSet(len(prefixSet.innerMap))

	for e1, zSet := range prefixSet.innerMap {
		if e1 == "" {
			continue
		}

		for z := range zSet.All() {
			es := z.GetSkuExternal().GetTags()

			intersection := ids.IntersectPrefixes(
				es,
				e,
			)

			exactMatch := intersection.Len() == 1 &&
				intersection.Any().Equals(e)

			if intersection.Len() == 0 && !exactMatch {
				continue
			}

			for _, e2 := range quiter.Elements(intersection) {
				out.Grouped.addPair(e2.String(), z)
			}
		}
	}

	return
}

func (prefixSet PrefixSet) Subset(
	e ids.Tag,
) (out Segments) {
	out.Ungrouped = makeObjSet()
	out.Grouped = MakePrefixSet(len(prefixSet.innerMap))

	e2 := catgut.MakeFromString(e.String())

	for e1, zSet := range prefixSet.innerMap {
		if e1 == "" {
			continue
		}

		for z := range zSet.All() {
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
		}
	}

	return
}
