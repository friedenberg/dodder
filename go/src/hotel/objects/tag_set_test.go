package objects

// func TestAddNormalized(t1 *testing.T) {
// 	t := &ui.T{T: t1}

// 	sut := MakeTagSetMutable(
// 		MustTag("project-2021-dodder-test"),
// 		MustTag("project-2021-dodder-ewwwwww"),
// 		MustTag("zz-archive-task-done"),
// 	)

// 	sutEx := CloneTagSet(sut)

// 	toAdd := MustTag("project-2021-dodder")

// 	AddNormalizedTag(sut, toAdd)

// 	if !quiter_set.Equals(sut, sutEx) {
// 		ui.TestPrintDiffString(
// 			t,
// 			quiter.StringCommaSeparated(sutEx),
// 			quiter.StringCommaSeparated(sut),
// 		)
// 	}
// }

// func TestAddNormalizedEmpty(t *testing.T) {
// 	sut := MakeTagSetMutable()
// 	toAdd := MustTag("project-2021-dodder")

// 	sutEx := MakeTagSetMutable(toAdd)

// 	AddNormalizedTag(sut, toAdd)

// 	if !quiter_set.Equals(sut, sutEx) {
// 		t.Errorf("expected %v, but got %v", sutEx, sut)
// 	}
// }

// func TestAddNormalizedFromEmptyBuild(t *testing.T) {
// 	toAdd := []Tag{
// 		MustTag("priority"),
// 		MustTag("priority-1"),
// 	}

// 	sut := MakeTagSetMutable()

// 	sutEx := MakeTagSetMutable(
// 		MustTag("priority-1"),
// 	)

// 	for _, e := range toAdd {
// 		AddNormalizedTag(sut, e)
// 	}

// 	if !quiter_set.Equals(sut, sutEx) {
// 		t.Errorf("expected %v, but got %v", sutEx, sut)
// 	}
// }

// func TestNormalize(t *testing.T) {
// 	type testEntry struct {
// 		ac TagSet
// 		ex TagSet
// 	}

// 	testEntries := map[string]testEntry{
// 		"removes all": {
// 			ac: MakeTagSetFromSlice(
// 				MustTag("project-2021-dodder"),
// 				MustTag("project-2021-dodder-test"),
// 				MustTag("project-2021-dodder-ewwwwww"),
// 				MustTag("project-archive-task-done"),
// 			),
// 			ex: MakeTagSetFromSlice(
// 				MustTag("project-2021-dodder-test"),
// 				MustTag("project-2021-dodder-ewwwwww"),
// 				MustTag("project-archive-task-done"),
// 			),
// 		},
// 		"removes non": {
// 			ac: MakeTagSetFromSlice(
// 				MustTag("project-2021-dodder"),
// 				MustTag("project-2021-dodder-test"),
// 				MustTag("project-2021-dodder-ewwwwww"),
// 				MustTag("zz-archive-task-done"),
// 			),
// 			ex: MakeTagSetFromSlice(
// 				MustTag("project-2021-dodder-test"),
// 				MustTag("project-2021-dodder-ewwwwww"),
// 				MustTag("zz-archive-task-done"),
// 			),
// 		},
// 		"removes right order": {
// 			ac: MakeTagSetFromSlice(
// 				MustTag("priority"),
// 				MustTag("priority-1"),
// 			),
// 			ex: MakeTagSetFromSlice(
// 				MustTag("priority-1"),
// 			),
// 		},
// 	}

// 	for d, te := range testEntries {
// 		t.Run(
// 			d,
// 			func(t1 *testing.T) {
// 				t := ui.T{T: t1}
// 				ac := collections_slice.Collect(te.ac.All())
// 				RemoveCommonPrefixes[TagStruct](&ac)

// 				expected := quiter.StringCommaSeparated(te.ex)
// 				actual := quiter.StringCommaSeparated(ac)

// 				t.AssertEqual(expected, actual)
// 			},
// 		)
// 	}
// }
