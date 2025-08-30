package tridex

import (
	"encoding/gob"
	"sort"
	"strings"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func init() {
	gob.Register(&Tridex{})
}

// TODO-P4 make generic
// TODO-P4 recycle nodes
// TODO-P4 confirm JSON structure is correct
// TODO switch to runes and rune readers as input
type Tridex struct {
	lock sync.RWMutex
	Root node
}

type node struct {
	Count            int
	Children         map[byte]node
	Value            string
	IsRoot           bool
	IncludesTerminus bool
}

func Make(vs ...string) (t interfaces.MutableTridex) {
	t = &Tridex{
		Root: node{
			Children: make(map[byte]node),
			IsRoot:   true,
		},
	}

	vs1 := make([]string, len(vs))
	copy(vs1, vs)

	sort.Slice(vs1, func(i, j int) bool { return len(vs1[i]) > len(vs1[j]) })

	for _, v := range vs1 {
		t.Add(v)
	}

	return
}

func (a *Tridex) MutableClone() (b interfaces.MutableTridex) {
	ui.TodoP4("improve the performance of this")
	ui.TodoP4("collections-copy")
	ui.TodoP4("collections-reset")
	ui.TodoP4("collections-recycle")

	a.lock.RLock()
	defer a.lock.RUnlock()

	b = &Tridex{
		Root: a.Root.Copy(),
	}

	return
}

func (t *Tridex) Len() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.Root.Count
}

func (t *Tridex) ContainsAbbreviation(v string) bool {
	return t.Contains(v)
}

func (t *Tridex) Contains(v string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.Root.Contains(v)
}

func (t *Tridex) ContainsExpansion(v string) bool {
	return t.ContainsExactly(v)
}

func (t *Tridex) ContainsExactly(v string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.Root.ContainsExactly(v)
}

func (t *Tridex) Abbreviate(v string) string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.Root.Abbreviate(v, 0)
}

func (t *Tridex) Expand(v string) string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	sb := &strings.Builder{}
	ok := t.Root.Expand(v, sb)

	if ok {
		return sb.String()
	} else {
		return v
	}
}

func (t *Tridex) Remove(v string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.Root.Remove(v)
}

func (t *Tridex) Add(v string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.Root.ContainsExactly(v) {
		return
	}

	t.Root.Add(v)
}

func (t *Tridex) EachString(f interfaces.FuncIter[string]) (err error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err = t.Root.Each(f, ""); err != nil {
		if errors.IsStopIteration(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	return
}

// TODO-P2 add Each and EachPtr methods
// func (t Tridex) GobEncode() (by []byte, err error) {
// 	bu := &bytes.Buffer{}
// 	enc := gob.NewEncoder(bu)
// 	err = enc.Encode(t.Root)
// 	by = bu.Bytes()
// 	return
// }

// func (t *Tridex) UnmarshalJSON(b []byte) error {
// 	bu := bytes.NewBuffer(b)
// 	dec := json.NewDecoder(bu)
// 	return dec.Decode(&t.Root)
// }

// func (t Tridex) MarshalJSON() (by []byte, err error) {
// 	bu := &bytes.Buffer{}
// 	enc := json.NewEncoder(bu)
// 	err = enc.Encode(t.Root)
// 	by = bu.Bytes()
// 	return
// }

// func (t *Tridex) GobDecode(b []byte) error {
// 	bu := bytes.NewBuffer(b)
// 	dec := gob.NewDecoder(bu)
// 	return dec.Decode(&t.Root)
// }
