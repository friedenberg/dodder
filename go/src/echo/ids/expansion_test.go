package ids

import (
	"reflect"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func stringSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestStringSliceUnequal(t1 *testing.T) {
	t := ui.T{T: t1}

	expected := []string{
		"this",
		"is",
		"a",
	}

	actual := []string{
		"this",
		"is",
		"a",
		"tag",
	}

	if stringSliceEquals(expected, actual) {
		t.Errorf("expected unequal slices")
	}
}

func TestStringSliceEquals(t1 *testing.T) {
	t := ui.T{T: t1}

	expected := []string{
		"this",
		"is",
		"a",
		"tag",
	}

	actual := []string{
		"this",
		"is",
		"a",
		"tag",
	}

	if !stringSliceEquals(expected, actual) {
		t.Errorf("expected equal slices")
	}
}

func TestExpansionAll(t1 *testing.T) {
	t := ui.T{T: t1}
	e := MustTag("this-is-a-tag")
	ex := MakeMutableTagSet()

	ExpandOneInto(
		e,
		MakeTag,
		expansion.ExpanderAll,
		ex,
	)

	expected := []string{
		"a",
		"a-tag",
		"is",
		"is-a-tag",
		"tag",
		"this",
		"this-is",
		"this-is-a",
		"this-is-a-tag",
	}

	actual := quiter.SortedStrings(ex)

	if !stringSliceEquals(actual, expected) {
		t.Errorf(
			"expanded tags don't match:\nexpected: %q\n  actual: %q",
			expected,
			actual,
		)
	}
}

func TestExpansionRight(t1 *testing.T) {
	t := ui.T{T: t1}

	e := MustTag("this-is-a-tag")
	ex := MakeMutableTagSet()

	ExpandOneInto(
		e,
		MakeTag,
		expansion.ExpanderRight,
		ex,
	)

	expected := []string{
		"this",
		"this-is",
		"this-is-a",
		"this-is-a-tag",
	}

	actual := quiter.SortedStrings[Tag](ex)

	if !stringSliceEquals(actual, expected) {
		t.Errorf(
			"expanded tags don't match:\nexpected: %q\n  actual: %q",
			expected,
			actual,
		)
	}
}

func TestExpansionRightTypeNone(t1 *testing.T) {
	t := ui.T{T: t1}
	e := MustType("md")

	actual := ExpandOneSlice(
		e,
		MakeType,
		expansion.ExpanderRight,
	)

	expected := []Type{
		MustType("md"),
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(
			"expanded tags don't match:\nexpected: %q\n  actual: %q",
			expected,
			actual,
		)
	}
}
