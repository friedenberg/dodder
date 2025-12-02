package objects

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	contents struct {
		// required to be exported for Gob's stupid illusions
		// TODO refactor this to use binary searches
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

// TODO switch to binary search
func (contents contents) ContainsKey(key string) bool {
	for id := range contents.Elements.All() {
		if id.GetKey().String() == key {
			return true
		}
	}

	return false
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

// TODO switch to binary search
func (contents contents) Get(key string) (SeqId, bool) {
	for id := range contents.Elements.All() {
		if id.GetKey().String() == key {
			return id.GetKey(), true
		}
	}

	return SeqId{}, false
}

func (contents contents) Key(id SeqId) string {
	return id.String()
}

// TODO sort
func (contents *contents) Add(id SeqId) error {
	if _, alreadyExists := contents.Get(id.String()); alreadyExists {
		return nil
	}

	contents.Elements.Append(containedObject{
		Lock: markl.MakeLockWith(id, nil),
	})

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
		contents.Elements.Delete(index, index+1)
	}

	return nil
}

func (contents *contents) Reset() {
	contents.Elements.Reset()
}
