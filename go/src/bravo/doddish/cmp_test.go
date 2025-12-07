package doddish

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

type cmpTestCase struct {
	left, right string
	expected    cmp.Result
}

func getCmpTestCases() []cmpTestCase {
	return []cmpTestCase{
		{
			left:     "tag",
			right:    "tag",
			expected: cmp.Equal,
		},
		{
			left:     "-tag",
			right:    "tag",
			expected: cmp.Less,
		},
		{
			left:     "zettel/id",
			right:    "tag",
			expected: cmp.Greater,
		},
	}
}

func makeSeqFromString(t *ui.T, input string) Seq {
	var scanner Scanner

	reader, repool := pool.GetStringReader(input)
	defer repool()

	scanner.Reset(reader)

	var index int

	var seq Seq

	for scanner.ScanDotAllowedInIdentifiers() {
		if index > 0 {
			t.Errorf("more than one seq in scanner")
		}

		seq = scanner.GetSeq()
		index++
	}

	if err := scanner.Error(); err != nil {
		t.AssertNoError(err)
	}

	return seq
}

func TestCmp(t1 *testing.T) {
	t := ui.T{T: t1}

	for _, testCase := range getCmpTestCases() {
		left := makeSeqFromString(&t, testCase.left)
		right := makeSeqFromString(&t, testCase.right)
		actual := SeqCompare(left, right)
		t.AssertEqual(testCase.expected, actual)
	}
}
