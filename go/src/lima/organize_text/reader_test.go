package organize_text

import (
	"os"
	"strings"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func makeZettelId(t *testing.T, v string) (k *ids.ObjectId) {
	var err error

	var h ids.ZettelId

	if err = h.Set(v); err != nil {
		t.Fatalf("%s", err)
	}

	return ids.MustObjectId(h)
}

func makeDescription(t *testing.T, v string) (b descriptions.Description) {
	var err error

	if err = b.Set(v); err != nil {
		t.Fatalf("%s", err)
	}

	return
}

func makeObjWithZettelIdAndDescription(
	t *testing.T,
	zettelId string,
	description string,
) (o *obj) {
	object := sku.MakeSkuType()
	object.GetSkuExternal().Metadata.Description = makeDescription(
		t,
		description,
	)

	o = &obj{
		sku: object,
	}

	o.GetSkuExternal().ObjectId.SetWithIdLike(makeZettelId(t, zettelId))

	return
}

func makeAssignmentLineReader() reader {
	return reader{
		options: Options{
			wasMade:       true,
			Config:        &repo_configs.TestDryRunOnly{},
			ObjectFactory: (&sku.ObjectFactory{}).SetDefaultsIfNecessary(),
			fmtBox: box_format.MakeBoxCheckedOut(
				string_format_writer.ColorOptions{},
				options_print.Options{},
				nil,
				ids.Abbr{},
				nil,
				nil,
				nil,
			),
		},
	}
}

func assertEqualObjects(t *ui.T, expected, actual Objects) {
	t = t.Skip(1)

	actual.Sort()
	expected.Sort()

	if len(actual) != len(expected) {
		t.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
	}

	for i := range actual {
		// actualObj, expectedObj := actual[i].External.GetSkuExternal(),
		// expected[i].External.GetSkuExternal()
		actualObj := sku.StringMetadataSansTai(actual[i].GetSkuExternal())
		expectedObj := sku.StringMetadataSansTai(expected[i].GetSkuExternal())

		if actualObj != expectedObj {
			t.Errorf("\nexpected: %#v\n  actual: %#v", expectedObj, actualObj)
		}
	}
}

func TestAssignmentLineReaderOneHeadingNoZettels(t1 *testing.T) {
	t := ui.T{T: t1}
	input := `# wow
    `

	sr := strings.NewReader(input)
	sub := makeAssignmentLineReader()

	n, err := sub.ReadFrom(sr)

	if n == 0 {
		t.Errorf("expected read amount to be greater than 0")
	}

	t.AssertNoError(err)

	{
		expected := ids.MakeTagSet(ids.MustTag("wow"))

		if len(sub.root.Children) < 1 {
			t.Fatalf("expected exactly 1 child")
		}

		actual := sub.root.Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(
			actual,
			expected,
		) {
			t.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}
}

func TestAssignmentLineReader2Heading2Zettels(t *testing.T) {
	t1 := ui.T{T: t}

	input := `# wow
- [one/wow] uno
- [dos/wow] two/wow
`

	sr := strings.NewReader(input)
	sub := makeAssignmentLineReader()

	n, err := sub.ReadFrom(sr)

	if n == 0 {
		t1.Errorf("expected read amount to be greater than 0")
	}

	if err != nil {
		t1.Fatalf("expected no error but got %q", err)
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("wow"))
		actual := sub.root.Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(
			actual,
			expected,
		) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "one/wow", "uno"))
		expected.Add(makeObjWithZettelIdAndDescription(t, "dos/wow", "two/wow"))

		actual := sub.root.Children[0].Objects

		assertEqualObjects(&t1, expected, actual)
	}

	if false {
		t := ui.T{T: t}
		var actualOut strings.Builder
		sut := Text{
			Options:    sub.options,
			Assignment: sub.root,
		}

		_, err := sut.WriteTo(&actualOut)
		t.AssertNoError(err)

		t.AssertEqual(input, actualOut.String())
	}
}

func TestAssignmentLineReader1_1Heading2_2Zettels(t1 *testing.T) {
	t := ui.T{T: t1}

	input := `# wow
## sub-wow
- [one/wow] uno
- [dos/wow] two/wow
`

	sr := strings.NewReader(input)
	sub := makeAssignmentLineReader()

	n, err := sub.ReadFrom(sr)

	if n == 0 {
		t.Errorf("expected read amount to be greater than 0")
	}

	if err != nil {
		t.Fatalf("expected no error but got %q", err)
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("wow"))
		actual := sub.root.Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(
			actual,
			expected,
		) {
			t.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("sub-wow"))

		if sub.root != sub.root.Children[0].Parent {
			t.Fatalf("%v, %v", sub.root, sub.root.Children[0].Parent)
		}

		l := len(sub.root.Children[0].Children)

		if l != 1 {
			t.Fatalf("\nexpected: %d\n  actual: %d", 1, l)
		}

		actual := sub.root.Children[0].Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(
			actual,
			expected,
		) {
			t.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t.T, "one/wow", "uno"))
		expected.Add(
			makeObjWithZettelIdAndDescription(t.T, "dos/wow", "two/wow"),
		)

		actual := sub.root.Children[0].Children[0].Objects

		assertEqualObjects(&t, expected, actual)
	}
}

func TestAssignmentLineReader2_1Heading2_2_2Zettels(t *testing.T) {
	t1 := ui.T{T: t}

	input := `# wow

- [one/wow] uno
- [dos/wow] two/wow

## sub-wow

- [three/wow] tres
- [four/wow] quatro

# cow

- [one/wow] uno
- [dos/wow] two/wow
`

	sr := strings.NewReader(input)
	sub := makeAssignmentLineReader()

	n, err := sub.ReadFrom(sr)

	if n == 0 {
		t1.Errorf("expected read amount to be greater than 0")
	}

	if err != nil {
		t1.Fatalf("expected no error but got %q", err)
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("wow"))
		actual := sub.root.Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(
			actual,
			expected,
		) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("sub-wow"))

		l := len(sub.root.Children[0].Children)
		if l != 1 {
			t1.Fatalf("\nexpected: %d\n  actual: %d", 1, l)
		}

		actual := sub.root.Children[0].Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(
			actual,
			expected,
		) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("cow"))
		actual := sub.root.Children[1].Transacted.Metadata.Tags

		if !ids.TagSetEquals(
			actual,
			expected,
		) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "one/wow", "uno"))
		expected.Add(makeObjWithZettelIdAndDescription(t, "dos/wow", "two/wow"))

		actual := sub.root.Children[0].Objects

		assertEqualObjects(&t1, expected, actual)
	}

	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "one/wow", "uno"))
		expected.Add(makeObjWithZettelIdAndDescription(t, "dos/wow", "two/wow"))

		actual := sub.root.Children[1].Objects

		assertEqualObjects(&t1, expected, actual)
	}
}

func TestAssignmentLineReader2_1Heading2_2_2ZettelsOffset(t *testing.T) {
	t1 := ui.T{T: t}

	input := `
- [one/wow] uno
- [dos/wow] two/wow

## sub-wow

- [three/wow] tres
- [four/wow] quatro

## sub-cow

- [one/wow] uno
- [dos/wow] two/wow
`

	sr := strings.NewReader(input)
	sub := makeAssignmentLineReader()

	n, err := sub.ReadFrom(sr)

	if n == 0 {
		t1.Errorf("expected read amount to be greater than 0")
	}

	if err != nil {
		t1.Fatalf("expected no error but got %q", err)
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("sub-wow"))
		actual := sub.root.Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(actual, expected) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := ids.MakeTagSet(ids.MustTag("sub-cow"))

		l := len(sub.root.Children)
		expLen := 2
		if l != expLen {
			t1.Fatalf("\nexpected: %d\n  actual: %d", expLen, l)
		}

		actual := sub.root.Children[1].Transacted.Metadata.Tags

		if !ids.TagSetEquals(actual, expected) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "four/wow", "quatro"))
		expected.Add(makeObjWithZettelIdAndDescription(t, "three/wow", "tres"))

		actual := sub.root.Children[0].Objects

		assertEqualObjects(&t1, expected, actual)
	}

	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "one/wow", "uno"))
		expected.Add(makeObjWithZettelIdAndDescription(t, "dos/wow", "two/wow"))

		actual := sub.root.Children[1].Objects

		assertEqualObjects(&t1, expected, actual)
	}
}

func TestAssignmentLineReaderBigCheese(t *testing.T) {
	t1 := ui.T{T: t}

	input := `# task
- [one/wow] uno
- [two/wow] dos/wow
## priority-1
### w-2022-07-09
- [three/wow] tres
###
- [four/wow] quatro
## priority-2
- [five/wow] cinco
- [six/wow] seis
`

	sr := strings.NewReader(input)
	sub := makeAssignmentLineReader()

	n, err := sub.ReadFrom(sr)

	if n == 0 {
		t1.Errorf("expected read amount to be greater than 0")
	}

	if err != nil {
		t1.Fatalf("expected no error but got %q", err)
	}

	// `# task
	{
		expected := ids.MakeTagSet(ids.MustTag("task"))
		actual := sub.root.Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(actual, expected) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	// - [one/wow] uno
	// - [two/wow] dos/wow
	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "one/wow", "uno"))
		expected.Add(makeObjWithZettelIdAndDescription(t, "two/wow", "dos/wow"))

		actual := sub.root.Children[0].Objects

		assertEqualObjects(&t1, expected, actual)
	}

	// ## priority-1
	{
		expected := ids.MakeTagSet(ids.MustTag("priority-1"))

		e := 2
		l := len(sub.root.Children[0].Children)
		if l != e {
			t1.Fatalf("\nexpected: %d\n  actual: %d", e, l)
		}

		actual := sub.root.Children[0].Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(actual, expected) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	// ### w-2022-07-09
	{
		expected := ids.MakeTagSet(ids.MustTag("w-2022-07-09"))
		actual := sub.root.Children[0].Children[0].Children[0].Transacted.Metadata.Tags

		if !ids.TagSetEquals(actual, expected) {
			t1.Errorf("\nexpected: %s\n  actual: %s", expected, actual)
		}
	}

	// - [three/wow] tres
	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "three/wow", "tres"))

		actual := sub.root.Children[0].Children[0].Children[0].Objects

		assertEqualObjects(&t1, expected, actual)
	}

	// ##
	// - [four/wow] quatro
	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "four/wow", "quatro"))

		actual := sub.root.Children[0].Children[0].Objects

		assertEqualObjects(&t1, expected, actual)
	}

	// ## priority-2
	// - [five/wow] cinco
	// - [six/wow] seis
	// `
	{
		expected := make(Objects, 0)
		expected.Add(makeObjWithZettelIdAndDescription(t, "five/wow", "cinco"))
		expected.Add(makeObjWithZettelIdAndDescription(t, "six/wow", "seis"))

		actual := sub.root.Children[0].Children[1].Objects

		assertEqualObjects(&t1, expected, actual)
	}
}
