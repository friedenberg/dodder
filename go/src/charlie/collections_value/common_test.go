package collections_value

import (
	"reflect"
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

func makeStringValues(vs ...string) (out []values.String) {
	out = make([]values.String, len(vs))

	for i, v := range vs {
		out[i] = values.MakeString(v)
	}

	return
}

func assertSet(
	t ui.T,
	sut interfaces.SetLike[values.String],
	vals []values.String,
) {
	t.Helper()

	// Len() int
	if sut.Len() != len(vals) {
		t.Fatalf("expected len %d but got %d", len(vals), sut.Len())
	}

	// Key(T) string
	{
		v := "wow"
		k := sut.Key(values.MakeString(v))

		if k != v {
			t.Fatalf("expected key %s but got %s", v, k)
		}
	}

	// Get(string) (T, bool)
	{
		for _, v := range vals {
			v1, ok := sut.Get(v.String())

			if !ok {
				t.Fatalf("expected sut to contain %s", v)
			}

			if v1 != v {
				t.Fatalf("expected %s but got %s", v, v1)
			}
		}
	}

	// ContainsKey(string) bool
	{
		for _, v := range vals {
			ok := sut.ContainsKey(v.String())

			if !ok {
				t.Fatalf("expected sut to contain %s", v)
			}
		}
	}

	{
		ex := vals
		ac := quiter.Elements[values.String](sut)

		sort.Slice(ac, func(i, j int) bool { return ac[i].Less(ac[j]) })

		if !reflect.DeepEqual(ex, ac) {
			t.Fatalf("expected %s but got %s", ex, ac)
		}
	}

	// Contains(T) bool
	for _, v := range vals {
		if !sut.Contains(v) {
			t.Fatalf("expected %s to contain %s", sut, v)
		}
	}

	// Copy
	{
		sutCopy := sut.CloneSetLike()

		if !quiter.SetEquals(sut, sutCopy) {
			t.Fatalf("expected copy to equal original")
		}
	}

	// MutableCopy
	{
		sutCopy := sut.CloneMutableSetLike()

		if !quiter.SetEquals[values.String](sut, sutCopy) {
			t.Fatalf("expected mutable copy to equal original")
		}

		sutCopy.Reset()

		if quiter.SetEquals[values.String](sut, sutCopy) {
			t.Fatalf("expected reset mutable copy to not equal original")
		}
	}

	// Each(WriterFunc[T]) error
	// EachKey(WriterFuncKey) error
}
