package errors

import "testing"

func TestWrapSkip(t *testing.T) {
	err := New("top level")
	err = WrapSkip(0, err)

	expected := "# TestWrapSkip\nmain_test.go:7: top level"

	if err.Error() != expected {
		t.Errorf("expected %q but got %q", expected, err)
	}
}
