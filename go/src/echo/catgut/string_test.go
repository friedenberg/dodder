package catgut

import (
	"fmt"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestMain(m *testing.M) {
	ui.SetTesting()
	m.Run()
}

type testCaseCompare struct {
	a, b     string
	expected int
}

func getTestCasesCompare() []testCaseCompare {
	return []testCaseCompare{
		{
			a:        "test",
			b:        "test",
			expected: 0,
		},
		{
			a:        "xest",
			b:        "test",
			expected: 1,
		},
		{
			a:        "",
			b:        "test",
			expected: -1,
		},
	}
}

func TestCompare(t1 *testing.T) {
	for _, tc := range getTestCasesCompare() {
		t1.Run(
			fmt.Sprintf("%#v", tc),
			func(t1 *testing.T) {
				t := ui.T{T: t1}

				a := MakeFromString(tc.a)
				b := MakeFromString(tc.b)

				actual := a.Compare(b)

				if actual != tc.expected {
					t.Errorf("expected %d but got %d", tc.expected, actual)
				}
			},
		)
	}
}

func getTestCasesComparePartial() []testCaseCompare {
	return []testCaseCompare{
		{
			a:        "test",
			b:        "test",
			expected: 0,
		},
		{
			a:        "tests",
			b:        "test",
			expected: 0,
		},
		{
			a:        "test",
			b:        "tests",
			expected: -1,
		},
		{
			a:        "",
			b:        "test",
			expected: -1,
		},
	}
}

func TestComparePartial(t1 *testing.T) {
	for _, tc := range getTestCasesComparePartial() {
		t1.Run(
			fmt.Sprintf("%#v", tc),
			func(t1 *testing.T) {
				t := ui.T{T: t1}

				a := MakeFromString(tc.a)
				b := MakeFromString(tc.b)

				actual := a.ComparePartial(b)

				if actual != tc.expected {
					t.Errorf("expected %d but got %d", tc.expected, actual)
				}
			},
		)
	}
}
