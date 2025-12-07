package objects

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	// TODO rename to maybe lockfile or lockedIds or id references or marklIds?
	contents struct {
		// required to be exported for Gob's stupid illusions
		// TODO refactor this to use binary searches
		TagCount int
		Elements collections_slice.Slice[containedObject]
	}
)

func (contents contents) Len() int {
	return contents.Elements.Len()
}

func (contents contents) All() interfaces.Seq[SeqId] {
	return func(yield func(SeqId) bool) {
		for id := range contents.Elements.All() {
			if !yield(id.GetKey()) {
				return
			}
		}
	}
}

func (contents contents) ContainsKey(key string) bool {
	_, ok := cmp.BinarySearchFuncIndex(
		contents.Elements,
		key,
		func(left containedObject, right string) cmp.Result {
			return cmp.CompareUTF8(
				left.GetKey().Seq.GetComparable(),
				cmp.ComparableString(right),
				false,
			)
		},
	)

	return ok
}

func (contents contents) getLock(key string) (IdLock, bool) {
	for id := range contents.Elements.All() {
		if id.GetKey().String() == key {
			return id.Lock, true
		}
	}

	return nil, false
}

func (contents contents) getLockMutable(key string) (IdLockMutable, bool) {
	for index := range contents.Elements {
		id := &contents.Elements[index]

		if id.GetKey().String() == key {
			return &id.Lock, true
		}
	}

	return nil, false
}

func (contents contents) Get(key string) (SeqId, bool) {
	id, ok := contents.get(key, false)
	return id.GetKey(), ok
}

func (contents contents) GetPartial(key string) (SeqId, bool) {
	id, ok := contents.get(key, true)
	return id.GetKey(), ok
}

func (contents contents) get(key string, partial bool) (containedObject, bool) {
	element, ok := cmp.BinarySearchFuncElement(
		contents.Elements,
		key,
		func(left containedObject, right string) cmp.Result {
			return cmp.CompareUTF8(
				left.GetKey().Seq.GetComparable(),
				cmp.ComparableString(right),
				partial,
			)
		},
	)

	return element, ok
}

func (contents contents) Key(id SeqId) string {
	return id.String()
}

func (contents *contents) Add(id SeqId) error {
	if _, alreadyExists := contents.Get(id.String()); alreadyExists {
		return nil
	}

	contents.Elements.Append(containedObject{
		Lock: markl.MakeLockWith(id, nil),
	})

	if id.Genre == genres.Tag {
		contents.TagCount++
	}

	contents.Elements.SortWithComparer(containedObjectCompareKey)

	return nil
}

func (contents *contents) DelKey(key string) error {
	var found bool
	var index int
	var id containedObject

	for index, id = range contents.Elements {
		if id.GetKey().String() == key {
			found = true
			break
		}
	}

	if found {
		if id.GetKey().Genre == genres.Tag {
			contents.TagCount--
		}

		contents.Elements.Delete(index, index+1)
	}

	return nil
}

func (contents *contents) Reset() {
	contents.TagCount = 0
	contents.Elements.Reset()
}

// TODO add optimized non-sorted path for binary decoding
func (contents *contents) addNormalizedTag(tag Tag) {
	seq := expansion.ExpandOneIntoIds[SeqId](
		tag.String(),
		expansion.ExpanderRight,
	)

	for id, err := range seq {
		errors.PanicIfError(err)
		errors.PanicIfError(contents.Add(id))
	}

	sorted := quiter.SortedValuesBy(
		contents.Elements,
		containedObjectCompareKey,
	)

	var lastId *containedObject

	for index := range sorted {
		id := &sorted[index]

		if index == 0 {
			// no need to do anything, this is the first
			lastId = id
			continue
		}

		tagString := id.Lock.GetKey().String()
		lastTagString := lastId.Lock.GetKey().String()

		switch {
		case strings.HasPrefix(lastTagString, tagString):
			continue

			// replace the shorter value with the longer value that contains the
			// shorter value
		case strings.HasPrefix(tagString, lastTagString):
			if lastId.Lock.Value.IsEmpty() {
				contents.DelKey(lastTagString)
			}

		default:
			// keep the tag
		}

		lastId = id
	}
}
