package errors

import (
	ConTeXT "context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"golang.org/x/xerrors"
)

var (
	errContextComplete  = New("context complete")
	errContextCancelled = New("context cancelled")
	errContextRetry     = New("context retry")
)

type ContextWithEnv[T any] struct {
	Context
	Env T
}

type Context interface {
	ConTeXT.Context

	Cause() error
	Continue() bool
	ContinueOrPanicOnDone()
	SetCancelOnSIGTERM()
	SetCancelOnSIGINT()
	SetCancelOnSIGHUP()
	SetCancelOnSignals(signals ...os.Signal)
	Run(f func(Context)) error

	// `After` runs a function after the context is complete (regardless of any
	// errors). `After`s are run in the reverse order of when they are called, like
	// defers but on a whole-program level.
	After(f func() error)
	AfterWithContext(f func(Context) error)

	// `Must` executes a function even if the context has been cancelled. If the
	// function returns an error, `Must` cancels the context and offers a heartbeat to
	// panic. It is meant for defers that must be executed, like closing files,
	// flushing buffers, releasing locks.
	Must(f func() error)
	MustWithContext(f func(Context) error)
	MustClose(closer io.Closer)
	MustFlush(flusher Flusher)
	Cancel()

	// TODO disambiguate between errors and exceptions
	CancelWithError(err error)
	CancelWithErrorAndFormat(err error, f string, values ...any)
	CancelWithErrorf(f string, values ...any)
	CancelWithBadRequestError(err error)
	CancelWithBadRequestf(f string, values ...any)
	CancelWithNotImplemented()
}

type RetryableContext interface {
	Context
	Retry()
}

type context struct {
	ConTeXT.Context
	funcCancel ConTeXT.CancelCauseFunc
	funcRun    func(Context)

	signals chan os.Signal

	lockRun       sync.Mutex
	lockConc      sync.Mutex
	doAfter       []FuncWithStackInfo
	doAfterErrors []error // TODO expose and use

	retriesDisabled bool
}

func MakeContextDefault() *context {
	return MakeContext(ConTeXT.Background())
}

func MakeContext(in ConTeXT.Context) *context {
	ctx, cancel := ConTeXT.WithCancelCause(in)

	return &context{
		Context:    ctx,
		funcCancel: cancel,
		signals:    make(chan os.Signal, 1),
	}
}

func (c *context) Cause() error {
	if err := ConTeXT.Cause(c.Context); err != nil {
		switch err {
		case errContextComplete, errContextCancelled:
			return nil

		default:
			return err
		}
	}

	return nil
}

func (c *context) Continue() bool {
	select {
	default:
		return true

	case <-c.Done():
		return false
	}
}

func (c *context) ContinueOrPanicOnDone() {
	if !c.Continue() {
		panic(errContextCancelled)
	}
}

func (c *context) SetCancelOnSIGTERM() {
	c.SetCancelOnSignals(syscall.SIGTERM)
}

func (c *context) SetCancelOnSIGINT() {
	c.SetCancelOnSignals(syscall.SIGINT)
}

func (c *context) SetCancelOnSIGHUP() {
	c.SetCancelOnSignals(syscall.SIGHUP)
}

func (c *context) SetCancelOnSignals(signals ...os.Signal) {
	signal.Notify(c.signals, signals...)
}

func (ctx *context) Run(funcRun func(Context)) error {
	if !ctx.lockRun.TryLock() {
		return ErrorWithStackf("Context.Run called before previous run completed.")
	}

	defer ctx.lockRun.Unlock()

	defer ctx.runAfter()

	ctx.funcRun = funcRun

	go func() {
		defer signal.Stop(ctx.signals)

		select {
		case <-ctx.Done():
		case sig := <-ctx.signals:
			ctx.cancel(Signal{Signal: sig})
		}
	}()

	for ctx.runRetry() {
	}

	return ctx.Cause()
}

func (ctx *context) runRetry() (shouldRetry bool) {
	defer func() {
		if r := recover(); r != nil {
			if r == errContextRetry {
				shouldRetry = true
				return
			}

			// TODO capture panic stack trace and add to custom error objects
			switch err := r.(type) {
			default:
				fmt.Printf("%s", debug.Stack())
				panic(r)

			case runtime.Error:
				fmt.Printf("%s", debug.Stack())
				panic(r)

			case error:
				ctx.cancel(err)
			}
		}
	}()

	ctx.funcRun(ctx)
	ctx.cancel(errContextComplete)

	return
}

func (ctx *context) runAfter() {
	for i := len(ctx.doAfter) - 1; i >= 0; i-- {
		doAfter := ctx.doAfter[i]
		err := doAfter.Func()
		if err != nil {
			ctx.doAfterErrors = append(
				ctx.doAfterErrors,
				doAfter.Wrap(err),
			)
		}
	}
}

func (ctx *context) Retry() {
	panic(errContextRetry)
}

func (ctx *context) cancel(err error) {
	var retryable Retryable

	if !ctx.retriesDisabled && As(err, &retryable) {
		retryable.Recover(ctx, err)
	} else {
		ctx.funcCancel(err)
	}
}

//go:noinline
func (c *context) after(skip int, f func() error) {
	c.lockConc.Lock()
	defer c.lockConc.Unlock()

	frame, _ := stack_frame.MakeFrame(skip + 1)

	c.doAfter = append(
		c.doAfter,
		FuncWithStackInfo{
			Func:  f,
			Frame: frame,
		},
	)
}

// `After` runs a function after the context is complete (regardless of any
// errors). `After`s are run in the reverse order of when they are called, like
// defers but on a whole-program level.
//
//go:noinline
func (c *context) After(f func() error) {
	c.after(1, f)
}

//go:noinline
func (c *context) AfterWithContext(f func(Context) error) {
	c.after(1, func() error { return f(c) })
}

// `Must` executes a function even if the context has been cancelled. If the
// function returns an error, `Must` cancels the context and offers a heartbeat to
// panic. It is meant for defers that must be executed, like closing files,
// flushing buffers, releasing locks.
func (c *context) Must(f func() error) {
	defer c.ContinueOrPanicOnDone()

	if err := f(); err != nil {
		c.cancel(WrapN(1, err))
	}
}

func (c *context) MustWithContext(f func(Context) error) {
	defer c.ContinueOrPanicOnDone()

	if err := f(c); err != nil {
		c.cancel(WrapN(1, err))
	}
}

func (c *context) MustClose(closer io.Closer) {
	c.Must(closer.Close)
}

func (c *context) MustFlush(flusher Flusher) {
	c.Must(flusher.Flush)
}

// TODO make this private and part of the run method
func (c *context) Cancel() {
	defer c.ContinueOrPanicOnDone()
	c.cancelWithoutPanic()
}

func (c *context) cancelWithoutPanic() {
	c.cancel(errContextCancelled)
}

func (c *context) CancelWithError(err error) {
	defer c.ContinueOrPanicOnDone()
	c.cancel(WrapN(1, err))
}

func (c *context) CancelWithErrorAndFormat(err error, f string, values ...any) {
	defer c.ContinueOrPanicOnDone()
	c.cancel(
		&stackWrapError{
			Frame: stack_frame.MustFrame(1),
			error: fmt.Errorf(f, values...),
			next:  WrapSkip(1, err),
		},
	)
}

func (c *context) CancelWithErrorf(f string, values ...any) {
	defer c.ContinueOrPanicOnDone()
	c.cancel(WrapSkip(1, fmt.Errorf(f, values...)))
}

func (c *context) CancelWithBadRequestError(err error) {
	defer c.ContinueOrPanicOnDone()
	c.cancel(&errBadRequestWrap{err})
}

func (c *context) CancelWithBadRequestf(f string, values ...any) {
	defer c.ContinueOrPanicOnDone()
	c.cancel(&errBadRequestWrap{xerrors.Errorf(f, values...)})
}

func (c *context) CancelWithNotImplemented() {
	defer c.ContinueOrPanicOnDone()
	c.cancel(ErrNotImplemented)
}
