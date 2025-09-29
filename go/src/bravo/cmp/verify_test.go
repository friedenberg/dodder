package cmp

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func TestLessorVerify(t1 *testing.T) {
	cmp := func(left, right int) Result {
		if left < right {
			return Less
		} else if left == right {
			return Equal
		} else {
			return Greater
		}
	}

	verifier := LesserVerify[int]([]interfaces.Lessor[int]{
		Lesser[int](cmp),
		Lesser[int](cmp),
	})

	if !verifier.Less(1, 2) {
		t1.Errorf("expected 1 to be less than 2")
	}

	if verifier.Less(2, 1) {
		t1.Errorf("expected 2 to not be less than 1")
	}
}
