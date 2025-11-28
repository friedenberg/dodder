package ids

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestAddNormalized(t1 *testing.T) {
	t := ui.T{T: t1}

	sut := MakeTagMutableSet(
		MustTag("project-2021-dodder-test"),
		MustTag("project-2021-dodder-ewwwwww"),
		MustTag("zz-archive-task-done"),
	)

	sutEx := CloneTagSet(sut)

	toAdd := MustTag("project-2021-dodder")

	AddNormalizedTag(sut, &toAdd)

	if !TagSetEquals(sut, sutEx) {
		t.NotEqual(sutEx, sut)
	}
}

func TestAddNormalizedEmpty(t *testing.T) {
	sut := MakeTagMutableSet()
	toAdd := MustTag("project-2021-dodder")

	sutEx := MakeTagMutableSet(toAdd)

	AddNormalizedTag(sut, &toAdd)

	if !TagSetEquals(sut, sutEx) {
		t.Errorf("expected %v, but got %v", sutEx, sut)
	}
}

func TestAddNormalizedFromEmptyBuild(t *testing.T) {
	toAdd := []Tag{
		MustTag("priority"),
		MustTag("priority-1"),
	}

	sut := MakeTagMutableSet()

	sutEx := MakeTagMutableSet(
		MustTag("priority-1"),
	)

	for i := range toAdd {
		e := &toAdd[i]
		AddNormalizedTag(sut, e)
	}

	if !TagSetEquals(sut, sutEx) {
		t.Errorf("expected %v, but got %v", sutEx, sut)
	}
}
