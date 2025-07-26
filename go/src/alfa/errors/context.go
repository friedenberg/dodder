package errors

import (
	ConTeXT "context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"golang.org/x/xerrors"
)

var errContextRetry = New("context retry")

type context struct {
	ConTeXT.Context

	stateLock sync.Mutex
	state     interfaces.ContextState

	stackFramesCancel []stack_frame.Frame
	lockCancel        sync.Mutex

	funcCancel ConTeXT.CancelCauseFunc
	funcRun    func(interfaces.Context)

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

func (ctx *context) Cause() error {
	if err := ConTeXT.Cause(ctx.Context); err != nil {
		switch err {
		case interfaces.ContextStateSucceeded, interfaces.ContextStateAborted:
			return nil

		default:
			return err
		}
	}

	return nil
}

func (ctx *context) CauseWithStackFrames() (error, []stack_frame.Frame) {
	if err := ctx.Cause(); err != nil {
		return err, ctx.stackFramesCancel
	} else {
		return nil, nil
	}
}

func (ctx *context) GetState() interfaces.ContextState {
	select {
	default:
		return interfaces.ContextStateStarted

	case <-ctx.Done():
		// TODO identify the right terminal state
		return interfaces.ContextStateSucceeded
	}
}

// TODO extricate from *context and turn into generic function
func (ctx *context) SetCancelOnSignals(signals ...os.Signal) {
	signal.Notify(ctx.signals, signals...)
}

func (ctx *context) Run(funcRun func(interfaces.Context)) error {
	if !ctx.lockRun.TryLock() {
		return ErrorWithStackf(
			"Context.Run called before previous run completed.",
		)
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

			switch err := r.(type) {
			default:
				ctx.captureCancelStackFramesIfNecessary(2, nil)
				ctx.cancel(xerrors.Errorf("context failed: %w", err))

			case string:
				{
					err := xerrors.Errorf("%s", err)
					ctx.captureCancelStackFramesIfNecessary(2, err)
					ctx.cancel(err)
				}

			case runtime.Error:
				ctx.captureCancelStackFramesIfNecessary(2, err)
				ctx.cancel(err)

			case error:
				ctx.cancel(err)
			}
		}
	}()

	ctx.funcRun(ctx)
	ctx.cancel(interfaces.ContextStateSucceeded)

	return
}

func (ctx *context) runAfter() {
	for i := len(ctx.doAfter) - 1; i >= 0; i-- {
		doAfter := ctx.doAfter[i]
		err := doAfter.FuncErr()
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
func (ctx *context) after(skip int, f func() error) {
	ctx.lockConc.Lock()
	defer ctx.lockConc.Unlock()

	frame, _ := stack_frame.MakeFrame(skip + 1)

	ctx.doAfter = append(
		ctx.doAfter,
		FuncWithStackInfo{
			FuncErr: f,
			Frame:   frame,
		},
	)
}

//go:noinline
func (ctx *context) After(f interfaces.FuncContext) {
	ctx.after(1, func() error { return f(ctx) })
}

// `Must` executes a function even if the context has been cancelled. If the
// function returns an error, `Must` cancels the context and offers a heartbeat
// to
// panic. It is meant for defers that must be executed, like closing files,
// flushing buffers, releasing locks.
func (ctx *context) Must(f interfaces.FuncContext) {
	defer ContextContinueOrPanic(ctx)

	if err := f(ctx); err != nil {
		ctx.cancel(WrapN(1, err))
	}
}

//go:noinline
func (ctx *context) Cancel(err error) {
	defer ContextContinueOrPanic(ctx)

	if err == nil {
		ctx.cancel(interfaces.ContextStateAborted)
		return
	}

	// TODO figure out why this needs to be 2
	ctx.captureCancelStackFramesIfNecessary(2, err)
	ctx.cancel(WrapN(1, err))
}

// TODO add interface for adding stack frames to the cancellation error
//
//go:noinline
func (ctx *context) captureCancelStackFramesIfNecessary(skip int, err error) {
	if !DebugBuild {
		return
	}

	switch err {
	case interfaces.ContextStateSucceeded, interfaces.ContextStateAborted:
		return
	}

	ctx.lockCancel.Lock()
	defer ctx.lockCancel.Unlock()

	defer func() {
		// if stack_frame.MakeFrames panics, we don't want that to take anything
		// else down
		recover()
	}()

	ctx.stackFramesCancel = stack_frame.MakeFrames(1+skip, 16)
}

//   __  __           _
//  |  \/  |_   _ ___| |_
//  | |\/| | | | / __| __|
//  | |  | | |_| \__ \ |_
//  |_|  |_|\__,_|___/\__|
//

func ContextMustClose(ctx interfaces.Context, closer io.Closer) {
	defer ContextContinueOrPanic(ctx)
	ctx.Must(MakeFuncContextFromFuncErr(closer.Close))
}

func ContextMustFlush(ctx interfaces.Context, flusher Flusher) {
	defer ContextContinueOrPanic(ctx)
	ctx.Must(MakeFuncContextFromFuncErr(flusher.Flush))
}

func ContextContinueOrPanic(ctx interfaces.Context) {
	if state := ctx.GetState(); state.IsComplete() {
		panic(state)
	}
}

//   ____  _                   _
//  / ___|(_) __ _ _ __   __ _| |___
//  \___ \| |/ _` | '_ \ / _` | / __|
//   ___) | | (_| | | | | (_| | \__ \
//  |____/|_|\__, |_| |_|\__,_|_|___/
//           |___/

func ContextSetCancelOnSIGTERM(ctx interfaces.Context) {
	ctx.SetCancelOnSignals(syscall.SIGTERM)
}

func ContextSetCancelOnSIGINT(ctx interfaces.Context) {
	ctx.SetCancelOnSignals(syscall.SIGINT)
}

func ContextSetCancelOnSIGHUP(ctx interfaces.Context) {
	ctx.SetCancelOnSignals(syscall.SIGHUP)
}

//    ____                     _
//   / ___|__ _ _ __   ___ ___| |___
//  | |   / _` | '_ \ / __/ _ \ / __|
//  | |__| (_| | | | | (_|  __/ \__ \
//   \____\__,_|_| |_|\___\___|_|___/
//

func ContextCancelWith499ClientClosedRequest(ctx interfaces.Context) {
	ctx.Cancel(Err499ClientClosedRequest)
}

func ContextCancelWithError(ctx interfaces.Context, err error) {
	defer ContextContinueOrPanic(ctx)
	ctx.Cancel(WrapN(1, err))
}

func ContextCancelWithErrorAndFormat(
	ctx interfaces.Context,
	err error,
	format string,
	values ...any,
) {
	defer ContextContinueOrPanic(ctx)
	ctx.Cancel(
		&stackWrapError{
			Frame: stack_frame.MustFrame(1),
			error: fmt.Errorf(format, values...),
			next:  WrapSkip(1, err),
		},
	)
}

func ContextCancelWithErrorf(
	ctx interfaces.Context,
	format string,
	values ...any,
) {
	defer ContextContinueOrPanic(ctx)
	ctx.Cancel(WrapSkip(1, fmt.Errorf(format, values...)))
}

func ContextCancelWithBadRequestError(ctx interfaces.Context, err error) {
	defer ContextContinueOrPanic(ctx)
	ctx.Cancel(&errBadRequestWrap{err})
}

func ContextCancelWithBadRequestf(
	ctx interfaces.Context,
	format string,
	values ...any,
) {
	defer ContextContinueOrPanic(ctx)
	ctx.Cancel(&errBadRequestWrap{xerrors.Errorf(format, values...)})
}

func CancelWithNotImplemented(ctx interfaces.Context) {
	defer ContextContinueOrPanic(ctx)
	ctx.Cancel(Err501NotImplemented)
}

//   ____                    _
//  |  _ \ _   _ _ __  _ __ (_)_ __   __ _
//  | |_) | | | | '_ \| '_ \| | '_ \ / _` |
//  |  _ <| |_| | | | | | | | | | | | (_| |
//  |_| \_\\__,_|_| |_|_| |_|_|_| |_|\__, |
//                                   |___/

func RunContextWithPrintTicker(
	context interfaces.Context,
	runFunc func(interfaces.Context),
	printFunc func(time.Time),
	duration time.Duration,
) (err error) {
	if err = context.Run(
		func(ctx interfaces.Context) {
			ticker := time.NewTicker(duration)
			ctx.After(MakeFuncContextFromFuncNil(ticker.Stop))

			go func() {
				for {
					select {
					case <-ctx.Done():
						return

					case t := <-ticker.C:
						printFunc(t)
					}
				}
			}()

			runFunc(ctx)
		},
	); err != nil {
		err = Wrap(err)
		return
	}

	return
}

func RunChildContextWithPrintTicker(
	parentContext interfaces.Context,
	runFunc func(interfaces.Context),
	printFunc func(time.Time),
	duration time.Duration,
) (err error) {
	return RunContextWithPrintTicker(
		MakeContext(parentContext),
		runFunc,
		printFunc,
		duration,
	)
}
