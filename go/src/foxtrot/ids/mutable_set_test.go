package ids

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestAddNormalized(t1 *testing.T) {
	t := ui.T{T: t1}

	sut := MakeTagSetMutable(
		MustTag("project-2021-dodder-test"),
		MustTag("project-2021-dodder-ewwwwww"),
		MustTag("zz-archive-task-done"),
	)

	sutEx := CloneTagSet(sut)

	toAdd := MustTag("project-2021-dodder")

	AddNormalizedTag(sut, toAdd)

	if !quiter_set.Equals(sut, sutEx) {
		t.NotEqual(sutEx, sut)
	}
}

func TestAddNormalizedEmpty(t *testing.T) {
	sut := MakeTagSetMutable()
	toAdd := MustTag("project-2021-dodder")

	sutEx := MakeTagSetMutable(toAdd)

	AddNormalizedTag(sut, toAdd)

	if !quiter_set.Equals(sut, sutEx) {
		t.Errorf("expected %v, but got %v", sutEx, sut)
	}
}

func TestAddNormalizedFromEmptyBuild(t *testing.T) {
	toAdd := []Tag{
		MustTag("priority"),
		MustTag("priority-1"),
	}

	sut := MakeTagSetMutable()

	sutEx := MakeTagSetMutable(
		MustTag("priority-1"),
	)

	for _, e := range toAdd {
		AddNormalizedTag(sut, e)
	}

	if !quiter_set.Equals(sut, sutEx) {
		t.Errorf("expected %v, but got %v", sutEx, sut)
	}
}
