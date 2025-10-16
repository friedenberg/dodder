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
}

func RunTestContext(
	t *testing.T,
	run func(*TestContext),
) {
	testContext := makeTestContext(t, errors.MakeContextDefault())

	if err := testContext.Context.Run(
		func(_ interfaces.Context) {
			run(testContext)
		},
	); err != nil {
		testContext.Skip(1).Fatalf("test context failed: %s", err)
	}
}

func makeTestContext(
	t *testing.T,
	ctx interfaces.Context,
) *TestContext {
	testContext := &TestContext{
		t: T{
			T: t,
		},
		Context: ctx,
	}

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
