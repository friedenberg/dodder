//go:build test

package ui

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var testStackFramePrefix = "    "

type T struct {
	*testing.T
	skip int
}

func (t *T) MakeStackInfo(skip int) (stackFrame stack_frame.Frame) {
	var pc uintptr
	ok := false
	pc, _, _, ok = runtime.Caller(skip + 1)

	if !ok {
		t.Fatal("failed to make stack info")
	}

	frames := runtime.CallersFrames([]uintptr{pc})

	frame, _ := frames.Next()
	stackFrame = stack_frame.MakeFrameFromRuntimeFrame(frame)
	stackFrame.Prefix = testStackFramePrefix

	return
}

func (t *T) SkipTest(args ...any) {
	if len(args) > 0 {
		t.ui(1, args...)
	}

	t.SkipNow()
}

func (t *T) Skip(skip int) *T {
	return &T{
		T:    t.T,
		skip: t.skip + skip,
	}
}

func (t *T) ui(skip int, args ...interface{}) {
	si := t.MakeStackInfo(t.skip + 1 + skip)
	args = append([]interface{}{si}, args...)
	fmt.Fprintln(os.Stderr, args...)
}

func (t *T) logf(skip int, format string, args ...interface{}) {
	si := t.MakeStackInfo(t.skip + 1 + skip).StringNoFunctionName()
	args = append([]interface{}{si}, args...)
	fmt.Fprintf(os.Stderr, "%s "+format+"\n", args...)
}

func (t *T) errorf(skip int, format string, args ...interface{}) {
	t.logf(skip+1, format, args...)
	t.Fail()
}

func (t *T) fatalf(skip int, format string, args ...interface{}) {
	t.logf(skip+1, format, args...)
	t.FailNow()
}

func (t *T) Log(args ...interface{}) {
	t.ui(1, args...)
}

func (t *T) Logf(format string, args ...interface{}) {
	t.logf(1, format, args...)
}

func (t *T) Errorf(format string, args ...interface{}) {
	t.Helper()
	t.errorf(1, format, args...)
}

func (t *T) Fatalf(format string, args ...interface{}) {
	t.Helper()
	t.fatalf(1, format, args...)
}

// TODO-P3 move to AssertNotEqual
func (t *T) NotEqual(a, b any) {
	t.errorf(1, "%s", cmp.Diff(a, b, cmpopts.IgnoreUnexported(a)))
}

func (t *T) AssertNotEqual(a, b any, o ...cmp.Option) {
	diff := cmp.Diff(a, b, o...)

	if diff == "" {
		return
	}

	t.errorf(1, "%s", diff)
}

func (t *T) AssertEqual(a, b any, o ...cmp.Option) {
	diff := cmp.Diff(a, b, o...)

	if diff == "" {
		return
	}

	t.errorf(1, "%s", diff)
}

func (t *T) AssertEqualStrings(a, b string) {
	t.Helper()

	if a == b {
		return
	}

	format := "\nexpected: %q\n  actual: %q"
	t.errorf(1, format, a, b)
}

func (t *T) AssertNoError(err error) {
	t.Helper()

	if err != nil {
		t.fatalf(1, "expected no error but got: %s", err)
	}
}

func (t *T) AssertEOF(err error) {
	t.Helper()

	if !errors.IsEOF(err) {
		t.fatalf(1, "expected EOF but got %q", err)
	}
}

func (t *T) AssertErrorEquals(expected, actual error) {
	t.Helper()

	if actual == nil {
		t.fatalf(1, "expected %q error but got none", expected)
	}

	if !errors.Is(actual, expected) {
		t.fatalf(1, "expected %q error but got %q", expected, actual)
	}
}

func (t *T) AssertError(err error) {
	t.Helper()

	if err == nil {
		t.fatalf(1, "expected an error but got none")
	}
}
