package ids

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestObjectId2IsVirtual(t1 *testing.T) {
	t := ui.T{T: t1}

	type expectedOutput struct {
		string
		isVirtual bool
	}

	type testCase struct {
		input string
		expectedOutput
	}

	testCases := []testCase{
		{
			input: "zz-site",
			expectedOutput: expectedOutput{
				string:    "%zz-site",
				isVirtual: false,
			},
		},
		{
			input: "%zz-site",
			expectedOutput: expectedOutput{
				string:    "%zz-site",
				isVirtual: true,
			},
		},
	}

	for _, testCase := range testCases {
		var sut objectId2
		err := sut.Set(testCase.input)
		t.AssertNoError(err)
		t.AssertEqual(testCase.isVirtual, sut.IsVirtual())
	}
}
