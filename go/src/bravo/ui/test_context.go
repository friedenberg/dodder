//go:build test

package ui

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type t = T

type TestContext struct {
	t

	interfaces.Context
	done <-chan struct{}
}

func MakeTestContext(t *testing.T) *TestContext {
	done := make(chan struct{})

	testContext := &TestContext{
		t:       T{T: t},
		Context: errors.MakeContextDefault(),
		done:    done,
	}

	go func() {
		if err := testContext.Context.Run(
			func(ctx interfaces.Context) {
				<-testContext.done
			},
		); err != nil {
			// TODO replay this `t.Fatalf` on the main go routine
			testContext.Fatalf("test contest failed: %s", err)
		}
	}()

	t.Cleanup(func() {
		defer func() {
			recover()
		}()

		testContext.Cancel(nil)
	},
	)

	return testContext
}

func (t *TestContext) Skip(skip int) *TestContext {
	return &TestContext{
		t:       *t.t.Skip(skip),
		Context: t.Context,
	}
}
