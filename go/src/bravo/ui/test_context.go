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

	Context interfaces.Context
	done    <-chan struct{}
}

func MakeTestContext(t *testing.T) *TestContext {
	return makeTestContext(t, errors.MakeContextDefault())
}

func makeTestContext(
	t *testing.T,
	ctx interfaces.Context,
) *TestContext {
	done := make(chan struct{})

	testContext := &TestContext{
		t: T{
			T: t,
		},
		Context: ctx,
		done:    done,
	}

	go func() {
		if err := testContext.Context.Run(
			func(ctx interfaces.Context) {
				<-testContext.done
			},
		); err != nil {
			// TODO replay this `t.Fatalf` on the main go routine
			testContext.Fatalf("test context failed: %s", err)
		}
	}()

	t.Cleanup(
		func() {
			defer func() {
				recover()
			}()

			testContext.Context.Cancel(nil)
		},
	)

	return testContext
}

func (testContext *TestContext) Skip(skip int) *TestContext {
	return &TestContext{
		t:       *testContext.t.Skip(skip),
		Context: testContext.Context,
	}
}

func (testContext *TestContext) Run(
	testCaseInfo TestCaseInfo,
	funk func(*TestContext),
) {
	description := getTestCaseDescription(testCaseInfo)

	testContext.T.Run(
		description,
		func(t1 *testing.T) {
			printTestCaseInfo(testCaseInfo, description)
			childContext := errors.MakeContext(testContext.Context)
			childTestContext := makeTestContext(t1, childContext)
			funk(childTestContext)
		},
	)
}
