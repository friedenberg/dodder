//go:build debug

package ui

import (
	"fmt"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

type TC struct {
	T
	stack_frame.Frame
}

func (t *TC) ui(args ...any) {
	args = append([]any{t.Frame}, args...)
	fmt.Fprintln(os.Stderr, args...)
}

func (t *TC) logf(format string, args ...any) {
	args = append([]any{t.Frame}, args...)
	fmt.Fprintf(os.Stderr, "%s"+format+"\n", args...)
}

func (t *TC) errorf(format string, args ...any) {
	t.logf(format, args...)
	t.Fail()
}

func (t *TC) fatalf(format string, args ...any) {
	t.logf(format, args...)
	t.FailNow()
}

func (t *TC) Log(args ...any) {
	t.ui(args...)
}

func (t *TC) Logf(format string, args ...any) {
	t.logf(format, args...)
}

func (t *TC) Errorf(format string, args ...any) {
	t.Helper()
	t.errorf(format, args...)
}

func (t *TC) Fatalf(format string, args ...any) {
	t.Helper()
	t.fatalf(format, args...)
}

// TODO-P3 move to AssertNotEqual
func (t *TC) NotEqual(a, b any) {
	format := "\nexpected: %q\n  actual: %q"
	t.errorf(format, a, b)
}

func (t *TC) AssertEqual(a, b any) {
	format := "\nexpected: %q\n  actual: %q"
	t.errorf(format, a, b)
}

func (t *TC) AssertEqualStrings(a, b string) {
	t.Helper()

	if a == b {
		return
	}

	format := "\nexpected: %q\n  actual: %q"
	t.errorf(format, a, b)
}

func (t *TC) AssertNoError(err error) {
	t.Helper()

	if err != nil {
		t.fatalf("expected no error but got %q", err)
	}
}

func (t *TC) AssertEOF(err error) {
	t.Helper()

	if !errors.IsEOF(err) {
		t.fatalf("expected EOF but got %q", err)
	}
}

func (t *TC) AssertErrorEquals(expected, actual error) {
	t.Helper()

	if actual == nil {
		t.fatalf("expected %q error but got none", expected)
	}

	if !errors.Is(actual, expected) {
		t.fatalf("expected %q error but got %q", expected, actual)
	}
}

func (t *TC) AssertError(err error) {
	t.Helper()

	if err == nil {
		t.fatalf("expected an error but got none")
	}
}
