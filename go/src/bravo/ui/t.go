//go:build test

package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var testStackFramePrefix = "    "

// TODO make this private and switch users over to MakeTestContext

type T struct {
	*testing.T
	skip int
}

//go:noinline
func (test *T) MakeStackInfo(skip int) (stackFrame stack_frame.Frame) {
	return stack_frame.MustFrame(skip + 1)
}

//go:noinline
func (test *T) SkipTest(args ...any) {
	if len(args) > 0 {
		test.ui(1, args...)
	}

	test.SkipNow()
}

func (test *T) Skip(skip int) *T {
	return &T{
		T:    test.T,
		skip: test.skip + skip,
	}
}

func (test *T) Run(testCaseInfo TestCaseInfo, funk func(*T)) {
	description := getTestCaseDescription(testCaseInfo)

	test.T.Run(
		description,
		func(t1 *testing.T) {
			printTestCaseInfo(testCaseInfo, description)
			funk(&T{T: t1})
		})
}

//   ___ ___
//  |_ _/ _ \
//   | | | | |
//   | | |_| |
//  |___\___/
//

//go:noinline
func (test *T) ui(skip int, args ...any) {
	stackFrame := test.MakeStackInfo(test.skip + 1 + skip)
	args = append([]any{stackFrame.StringNoFunctionName()}, args...)
	fmt.Fprintln(os.Stderr, args...)
}

//go:noinline
func (test *T) logf(skip int, format string, args ...any) {
	stackFrame := test.MakeStackInfo(test.skip + 1 + skip)
	args = append([]any{stackFrame.StringNoFunctionName()}, args...)
	fmt.Fprintf(os.Stderr, "%s "+format+"\n", args...)
}

//go:noinline
func (test *T) errorf(skip int, format string, args ...any) {
	test.logf(skip+1, format, args...)
	test.Fail()
}

//go:noinline
func (test *T) fatalf(skip int, format string, args ...any) {
	test.logf(skip+1, format, args...)
	test.FailNow()
}

//go:noinline
func (test *T) Log(args ...any) {
	test.ui(1, args...)
}

//go:noinline
func (test *T) Logf(format string, args ...any) {
	test.logf(1, format, args...)
}

//go:noinline
func (test *T) Errorf(format string, args ...any) {
	test.Helper()
	test.errorf(1, format, args...)
}

//go:noinline
func (test *T) Fatalf(format string, args ...any) {
	test.Helper()
	test.fatalf(1, format, args...)
}

//      _                      _
//     / \   ___ ___  ___ _ __| |_ ___
//    / _ \ / __/ __|/ _ \ '__| __/ __|
//   / ___ \\__ \__ \  __/ |  | |_\__ \
//  /_/   \_\___/___/\___|_|   \__|___/
//

// TODO-P3 move to AssertNotEqual
//
//go:noinline
func (test *T) NotEqual(a, b any) {
	test.errorf(1, "%s", cmp.Diff(a, b, cmpopts.IgnoreUnexported(a)))
}

//go:noinline
func (test *T) AssertNotEqual(a, b any, o ...cmp.Option) {
	diff := cmp.Diff(a, b, o...)

	if diff == "" {
		return
	}

	test.errorf(1, "%s", diff)
}

//go:noinline
func (test *T) AssertEqual(a, b any, o ...cmp.Option) {
	diff := cmp.Diff(a, b, o...)

	if diff == "" {
		return
	}

	test.errorf(1, "%s", diff)
}

//go:noinline
func (test *T) AssertEqualStrings(a, b string) {
	test.Helper()

	if a == b {
		return
	}

	format := "string equality failed\n=== expected ===\n%s\n=== actual ===\n%s"
	test.errorf(1, format, a, b)
}

//go:noinline
func (test *T) AssertNoError(err error) {
	test.Helper()

	if err != nil {
		test.fatalf(1, "expected no error but got: %s", err)
	}
}

//go:noinline
func (test *T) AssertEOF(err error) {
	test.Helper()

	if err != io.EOF {
		test.fatalf(1, "expected EOF but got %q", err)
	}
}

//go:noinline
func (test *T) AssertErrorEquals(expected, actual error) {
	test.Helper()

	if actual == nil {
		test.fatalf(1, "expected %q error but got none", expected)
	}

	if !errors.Is(actual, expected) {
		test.fatalf(1, "expected %q error but got %q", expected, actual)
	}
}

//go:noinline
func (test *T) AssertError(err error) {
	test.Helper()

	if err == nil {
		test.fatalf(1, "expected an error but got none")
	}
}
