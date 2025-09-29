package heap

import (
	"reflect"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/cmp"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

func TestMerge(t1 *testing.T) {
	t := ui.T{T: t1}

	eql := values.IntEqualer{}
	llr := values.IntLessor{}

	els := []*values.Int{
		values.MakeInt(1),
		values.MakeInt(0),
		values.MakeInt(3),
		values.MakeInt(4),
		values.MakeInt(2),
	}

	otherStream := MakeHeapFromSliceUnsorted(
		eql,
		llr,
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

	sut := MakeHeapFromSliceUnsorted(
		eql,
		llr,
		values.IntResetter{},
		els,
	)

	actual := make([]*values.Int, 0)

	seq := quiter.MergeSequences(
		quiter.MakeSeqErrorFromSeq(sut.All()),
		quiter.MakeSeqErrorFromSeq(otherStream.All()),
		func(left, right *values.Int) cmp.Result {
			if left.Less(*right) {
				return cmp.Less
			} else if left.Equals(*right) {
				return cmp.Equal
			} else {
				return cmp.Greater
			}
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

	eql := values.IntEqualer{}
	llr := values.IntLessor{}

	els := []*values.Int{
		values.MakeInt(1),
		values.MakeInt(0),
		values.MakeInt(3),
		values.MakeInt(4),
		values.MakeInt(2),
	}

	otherStream := MakeHeapFromSliceUnsorted[values.Int, *values.Int](
		eql,
		llr,
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

	sut := MakeHeapFromSliceUnsorted[values.Int, *values.Int](
		eql,
		llr,
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
