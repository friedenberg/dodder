package ids

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/collections_delta"
)

func TestNormalize(t *testing.T) {
	type testEntry struct {
		ac TagSet
		ex TagSet
	}

	testEntries := map[string]testEntry{
		"removes all": {
			ac: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("project-archive-task-done"),
			),
			ex: MakeTagSetFromSlice(
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("project-archive-task-done"),
			),
		},
		"removes non": {
			ac: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("zz-archive-task-done"),
			),
			ex: MakeTagSetFromSlice(
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("zz-archive-task-done"),
			),
		},
		"removes right order": {
			ac: MakeTagSetFromSlice(
				MustTag("priority"),
				MustTag("priority-1"),
			),
			ex: MakeTagSetFromSlice(
				MustTag("priority-1"),
			),
		},
	}

	for d, te := range testEntries {
		t.Run(
			d,
			func(t1 *testing.T) {
				t := ui.T{T: t1}
				ac := WithRemovedCommonPrefixes(te.ac)

				if !TagSetEquals(ac, te.ex) {
					t.Errorf(
						"removing prefixes doesn't match:\nexpected: %q\n  actual: %q",
						te.ex,
						ac,
					)
				}
			},
		)
	}
}

func TestRemovePrefixes(t *testing.T) {
	type testEntry struct {
		ac     TagSet
		ex     TagSet
		prefix string
	}

	testEntries := map[string]testEntry{
		"removes all": {
			ac: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("project-archive-task-done"),
			),
			ex:     MakeTagSetFromSlice(),
			prefix: "project",
		},
		"removes non": {
			ac: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("zz-archive-task-done"),
			),
			ex: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("zz-archive-task-done"),
			),
			prefix: "xx",
		},
		"removes one": {
			ac: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("zz-archive-task-done"),
			),
			ex: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
			),
			prefix: "zz",
		},
		"removes most": {
			ac: MakeTagSetFromSlice(
				MustTag("project-2021-dodder"),
				MustTag("project-2021-dodder-test"),
				MustTag("project-2021-dodder-ewwwwww"),
				MustTag("zz-archive-task-done"),
			),
			ex: MakeTagSetFromSlice(
				MustTag("zz-archive-task-done"),
			),
			prefix: "project",
		},
	}

	for d, te := range testEntries {
		t.Run(
			d,
			func(t *testing.T) {
				assertSetRemovesPrefixes(t, te.ac, te.ex, te.prefix)
			},
		)
	}
}

func TestExpandedRight(t *testing.T) {
	s := MakeTagSetFromSlice(
		MustTag("project-2021-dodder"),
		MustTag("zz-archive-task-done"),
	)

	ex := Expanded(s, expansion.ExpanderRight)

	expected := []string{
		"project",
		"project-2021",
		"project-2021-dodder",
		"zz",
		"zz-archive",
		"zz-archive-task",
		"zz-archive-task-done",
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

func TestPrefixIntersection(t *testing.T) {
	s := MakeTagSetFromSlice(
		MustTag("project-2021-dodder"),
		MustTag("zz-archive-task-done"),
	)

	ex := IntersectPrefixes(s, MustTag("project"))

	expected := []string{
		"project-2021-dodder",
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

// func TestExpansionRight(t *testing.T) {
// 	e := MustTag("this-is-a-tag")
// 	ex := e.Expanded(ExpanderRight{})
// 	expected := []string{
// 		"this",
// 		"this-is",
// 		"this-is-a",
// 		"this-is-a-tag",
// 	}

// 	actual := ex.SortedString()

// 	if !stringSliceEquals(actual, expected) {
// 		t.Errorf(
// 			"expanded tags don't match:\nexpected: %q\n  actual: %q",
// 			expected,
// 			actual,
// 		)
// 	}
// }

func TestDelta1(t *testing.T) {
	from := MakeTagSetFromSlice(
		MustTag("project-2021-dodder"),
		MustTag("task-todo"),
	)

	to := MakeTagSetFromSlice(
		MustTag("project-2021-dodder"),
		MustTag("zz-archive-task-done"),
	)

	d := collections_delta.MakeSetDelta(from, to)

	c_expected := MakeTagSetFromSlice(
		MustTag("zz-archive-task-done"),
	)

	if !quiter_set.Equals(c_expected, d.GetAdded()) {
		t.Errorf("expected\n%s\nactual:\n%s", c_expected, d.GetAdded())
	}

	d_expected := MakeTagSetFromSlice(
		MustTag("task-todo"),
	)

	if !quiter_set.Equals(d_expected, d.GetRemoved()) {
		t.Errorf("expected\n%s\nactual:\n%s", d_expected, d.GetRemoved())
	}
}
