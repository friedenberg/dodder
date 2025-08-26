package ids

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var (
	tagPool     interfaces.Pool[Tag, *Tag]
	tagPoolOnce sync.Once
)

type tagResetter struct{}

func (tagResetter) Reset(e *Tag) {
	e.value = ""
	e.virtual = false
	e.dependentLeaf = false
}

func (tagResetter) ResetWith(a, b *Tag) {
	a.value = b.value
	a.virtual = b.virtual
	a.dependentLeaf = b.dependentLeaf
}

type tag2Resetter struct{}

func (tag2Resetter) Reset(e *tag2) {
	e.value.Reset()
	e.virtual = false
	e.dependentLeaf = false
}

func (tag2Resetter) ResetWith(a, b *tag2) {
	b.value.CopyTo(a.value)
	a.virtual = b.virtual
	a.dependentLeaf = b.dependentLeaf
}

var (
	tagPtrMapPool     interfaces.PoolValue[map[string]*Tag]
	tagPtrMapPoolOnce sync.Once
)

func GetTagPool() interfaces.Pool[Tag, *Tag] {
	tagPoolOnce.Do(
		func() {
			tagPool = pool.Make(
				func() *Tag {
					e := &Tag{}
					e.init()
					return e
				},
				TagResetter.Reset,
			)
		},
	)

	return tagPool
}
