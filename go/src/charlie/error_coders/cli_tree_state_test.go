package error_coders

import (
	"strings"
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

type testCaseCLITreeState struct {
	ui.TestCaseInfo
	input    error
	expected string
}

func TestCLITreeForwards(t *testing.T) {
	tc := ui.MakeTestContext(t)

	type testCase = testCaseCLITreeState

	testCases := []testCase{
		{
			TestCaseInfo: ui.MakeTestCaseInfo("error group three"),
			input: errors.Group{
				errors.New("one"),
				errors.New("two"),
				errors.New("three"),
			},
			expected: `errors.Group: 3 errors
├── one
├── two
└── three
`,
		},
		{
			TestCaseInfo: ui.MakeTestCaseInfo(
				"error group three with nested child",
			),
			input: errors.Group{
				errors.New("one"),
				errors.New("two"),
				errors.Group{
					errors.New("three"),
				},
			},
			expected: `errors.Group: 3 errors
├── one
├── two
└── errors.Group: 1 errors
    └── three
`,
		},
		{
			TestCaseInfo: ui.MakeTestCaseInfo(
				"error group three with double nested child",
			),
			input: errors.Group{
				errors.New("one"),
				errors.New("two"),
				errors.Group{
					errors.Err501NotImplemented.Wrap(errors.New("inner")),
				},
			},
			expected: `errors.Group: 3 errors
├── one
├── two
└── errors.Group: 1 errors
    └── errors.HTTP: 501 Not Implemented
        └── inner
`,
		},
	}

	for _, testCase := range testCases {
		tc.Run(
			testCase,
			func(t *ui.TestContext) {
				var stringBuilder strings.Builder

				bufferedWriter, repool := pool.GetBufferedWriter(&stringBuilder)
				defer repool()

				coder := cliTreeState{
					bufferedWriter: bufferedWriter,
				}

				{
					err := coder.encode(testCase.input)

					if coder.bytesWritten == 0 {
						t.Errorf("expected non-zero bytes written")
					}

					t.AssertNoError(err)
				}

				actual := stringBuilder.String()

				t.AssertEqualStrings(testCase.expected, actual)
			},
		)
	}
}
