package heap

import (
	"reflect"
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/cmp"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

func makeTestFuncCmp(_ ui.T) cmp.Func[*values.Int] {
	return cmp.MakeFuncFromEqualerAndLessor3LessFirst(
		values.IntEqualer{},
		values.IntLessor{},
	)
}

type mergeTestCase[ELEMENT any, ELEMENT_PTR interfaces.Ptr[ELEMENT]] struct {
	info     ui.TestCaseInfo
	cmp      cmp.Func[ELEMENT_PTR]
	resetter interfaces.Resetter[ELEMENT_PTR]
	left     []ELEMENT_PTR
	right    []ELEMENT_PTR
	expected []ELEMENT_PTR
}

func (testCase mergeTestCase[ELEMENT, ELEMENT_PTR]) Test(t *ui.T) {
	left := MakeNewHeapFromSliceUnsorted(
		testCase.cmp,
		testCase.resetter,
		testCase.left,
	)

	right := MakeNewHeapFromSliceUnsorted(
		testCase.cmp,
		testCase.resetter,
		testCase.right,
	)

	seq := quiter.MergeSeqErrorLeft(
		quiter.MakeSeqErrorFromSeq(left.All()),
		quiter.MakeSeqErrorFromSeq(right.All()),
		testCase.cmp,
	)

	actual, err := quiter.CollectError(seq)
	t.AssertNoError(err)

	if !reflect.DeepEqual(testCase.expected, actual) {
		t.Skip(1).Errorf("expected %q but got %q", testCase.expected, actual)
	}
}

func TestMerge(t1 *testing.T) {
	t := ui.T{T: t1}

	funcCmp := makeTestFuncCmp(t)

	testCases := []ui.TestCase[mergeTestCase[values.Int, *values.Int]]{
		ui.MakeTestCase(
			"both empty",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left:     []*values.Int{},
				right:    []*values.Int{},
				expected: []*values.Int{},
			},
		),
		ui.MakeTestCase(
			"disjunct",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left:     []*values.Int{values.MakeInt(0)},
				right:    []*values.Int{values.MakeInt(1)},
				expected: []*values.Int{
					values.MakeInt(0),
					values.MakeInt(1),
				},
			},
		),
		ui.MakeTestCase(
			"left is a copy of right",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left:     []*values.Int{values.MakeInt(0)},
				right:    []*values.Int{values.MakeInt(0)},
				expected: []*values.Int{
					values.MakeInt(0),
				},
			},
		),
		ui.MakeTestCase(
			"left is a copy of right big edition",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left: []*values.Int{
					values.MakeInt(1),
					values.MakeInt(0),
					values.MakeInt(3),
					values.MakeInt(4),
					values.MakeInt(2),
				},
				right: []*values.Int{
					values.MakeInt(1),
					values.MakeInt(0),
					values.MakeInt(3),
					values.MakeInt(4),
					values.MakeInt(2),
				},
				expected: []*values.Int{
					values.MakeInt(0),
					values.MakeInt(1),
					values.MakeInt(2),
					values.MakeInt(3),
					values.MakeInt(4),
				},
			},
		),
		ui.MakeTestCase(
			"left overlaps right by one",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left: []*values.Int{
					values.MakeInt(1),
					values.MakeInt(0),
					values.MakeInt(3),
					values.MakeInt(4),
					values.MakeInt(2),
				},
				right: []*values.Int{
					values.MakeInt(8),
					values.MakeInt(9),
					values.MakeInt(3),
					values.MakeInt(7),
					values.MakeInt(6),
				},
				expected: []*values.Int{
					values.MakeInt(0),
					values.MakeInt(1),
					values.MakeInt(2),
					values.MakeInt(3),
					values.MakeInt(4),
					values.MakeInt(6),
					values.MakeInt(7),
					values.MakeInt(8),
					values.MakeInt(9),
				},
			},
		),
		ui.MakeTestCase(
			"completely interlaced",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left: []*values.Int{
					values.MakeInt(0),
					values.MakeInt(2),
					values.MakeInt(4),
					values.MakeInt(6),
					values.MakeInt(8),
				},
				right: []*values.Int{
					values.MakeInt(1),
					values.MakeInt(3),
					values.MakeInt(5),
					values.MakeInt(7),
					values.MakeInt(9),
				},
				expected: []*values.Int{
					values.MakeInt(0),
					values.MakeInt(1),
					values.MakeInt(2),
					values.MakeInt(3),
					values.MakeInt(4),
					values.MakeInt(5),
					values.MakeInt(6),
					values.MakeInt(7),
					values.MakeInt(8),
					values.MakeInt(9),
				},
			},
		),
		ui.MakeTestCase(
			"left empty",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left:     []*values.Int{},
				right: []*values.Int{
					values.MakeInt(8),
					values.MakeInt(9),
					values.MakeInt(3),
					values.MakeInt(7),
					values.MakeInt(6),
				},
				expected: []*values.Int{
					values.MakeInt(3),
					values.MakeInt(6),
					values.MakeInt(7),
					values.MakeInt(8),
					values.MakeInt(9),
				},
			},
		),
		ui.MakeTestCase(
			"right empty",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left: []*values.Int{
					values.MakeInt(1),
					values.MakeInt(0),
					values.MakeInt(3),
					values.MakeInt(4),
					values.MakeInt(2),
				},
				right: []*values.Int{},
				expected: []*values.Int{
					values.MakeInt(0),
					values.MakeInt(1),
					values.MakeInt(2),
					values.MakeInt(3),
					values.MakeInt(4),
				},
			},
		),
		ui.MakeTestCase(
			"both empty",
			mergeTestCase[values.Int, *values.Int]{
				cmp:      funcCmp,
				resetter: values.IntResetter{},
				left:     []*values.Int{},
				right:    []*values.Int{},
				expected: []*values.Int{},
			},
		),
	}

	for _, testCase := range testCases {
		t.Run(
			testCase,
			func(t *ui.T) {
				testCase.GetBlob().Test(t)
			},
		)
	}
}

func TestMergeEqual(t1 *testing.T) {
	t := ui.T{T: t1}

	funcCmp := makeTestFuncCmp(t)

	left := MakeNewHeapFromSliceUnsorted(
		funcCmp,
		values.IntResetter{},
		[]*values.Int{
			values.MakeInt(1),
		},
	)

	right := MakeNewHeapFromSliceUnsorted(
		funcCmp,
		values.IntResetter{},
		[]*values.Int{
			values.MakeInt(1),
		},
	)

	expected := []*values.Int{
		values.MakeInt(1),
	}

	actual := make([]*values.Int, 0)

	seq := quiter.MergeSeqErrorLeft(
		quiter.MakeSeqErrorFromSeq(left.All()),
		quiter.MakeSeqErrorFromSeq(right.All()),
		func(left, right *values.Int) cmp.Result {
			return cmp.Ordered(left.Int(), right.Int())
		},
	)

	for element, errIter := range seq {
		t.AssertNoError(errIter)
		actual = append(actual, element)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %q but got %q", expected, actual)
	}
}

func TestMergeAndRestore(t1 *testing.T) {
	t := ui.T{T: t1}

	funcCmp := cmp.MakeFuncFromEqualerAndLessor3LessFirst(
		values.IntEqualer{},
		values.IntLessor{},
	)

	els := []*values.Int{
		values.MakeInt(1),
		values.MakeInt(0),
		values.MakeInt(3),
		values.MakeInt(4),
		values.MakeInt(2),
	}

	otherStream := MakeNewHeapFromSliceUnsorted(
		funcCmp,
		values.IntResetter{},
		[]*values.Int{
			values.MakeInt(8),
			values.MakeInt(9),
			values.MakeInt(3),
			values.MakeInt(7),
			values.MakeInt(6),
		},
	)

	expected := []*values.Int{
		values.MakeInt(0),
		values.MakeInt(1),
		values.MakeInt(2),
		values.MakeInt(3),
		values.MakeInt(4),
		values.MakeInt(6),
		values.MakeInt(7),
		values.MakeInt(8),
		values.MakeInt(9),
	}

	sut := MakeNewHeapFromSliceUnsorted(
		funcCmp,
		values.IntResetter{},
		els,
	)

	actual := make([]*values.Int, 0)

	err := MergeHeapAndRestore(
		sut,
		otherStream.PopOrErrStopIteration,
		func(v *values.Int) (err error) {
			actual = append(actual, v)
			return err
		},
	)
	if err != nil {
		t.AssertNoError(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %q but got %q", expected, actual)
	}
}
