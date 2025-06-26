package errors

import (
	ConTeXT "context"
	"fmt"
	"syscall"
	"testing"
)

func TestContextCancelled(t *testing.T) {
	ctx := MakeContext(ConTeXT.Background())

	var must1, must2, after1 bool

	if err := ctx.Run(
		func(ctx Context) {
			didPanic := false

			defer func() {
				t.Log("defer1")

				if r := recover(); r != nil {
					t.Log("recover")
					didPanic = true

					if r != errContextCancelled {
						t.Errorf("expected recover to be %q", errContextCancelled)
					}
				}
			}()

			defer ctx.Must(func() error {
				t.Log("must1")
				must1 = true
				return nil
			})

			defer ctx.Must(func() error {
				t.Log("must2")
				must2 = true
				return nil
			})

			ctx.After(func() error {
				after1 = true
				return nil
			})

			ctx.Cancel()
			ctx.ContinueOrPanicOnDone()

			t.Errorf("expected to not get here")

			if !didPanic {
				t.Errorf("expected to panic")
			}
		},
	); err != nil {
		t.Errorf("expected no error but got: %s", err)
	}

	if !must1 || !must2 || !after1 {
		t.Errorf("expected all must and after functions to execute")
	}
}

type errTestRecover struct{}

func (errTestRecover) Error() string {
	return "test recover error"
}

func (err errTestRecover) GetRetryableError() Retryable {
	return err
}

func (errTestRecover) Recover(ctx RetryableContext, in error) {
	ctx.Retry()
}

func TestContextCancelledRetry(t *testing.T) {
	ctx := MakeContext(ConTeXT.Background())

	tryCount := 0

	if err := ctx.Run(
		func(ctx Context) {
			fmt.Printf("%d\n", tryCount)
			if tryCount == 0 {
				tryCount++
				ctx.CancelWithError(errTestRecover{})
			}

			tryCount++
		},
	); err != nil {
		t.Errorf("expected no error but got: %s", err)
	}

	if tryCount != 2 {
		t.Errorf("expected try count 2 but got: %d", tryCount)
	}
}

func TestContextSignal(t *testing.T) {
	ctx := MakeContext(ConTeXT.Background())
	ctx.SetCancelOnSIGHUP()

	cont := make(chan struct{})

	go func() {
		if err := ctx.Run(
			func(ctx Context) {
				child := MakeContext(ctx)

				if err := child.Run(
					func(ctx Context) {
						<-ctx.Done()
						cont <- struct{}{}
					},
				); err == nil {
					t.Errorf("expected signal error but got none")
				}
			},
		); err == nil {
			t.Errorf("expected signal error but got none")
		}
	}()

	ctx.signals <- syscall.SIGHUP
	<-cont
}
