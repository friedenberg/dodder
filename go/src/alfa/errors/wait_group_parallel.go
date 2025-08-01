package errors

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

func MakeWaitGroupParallel() WaitGroup {
	waitGroup := &waitGroupParallel{
		lock:         &sync.Mutex{},
		inner:        &sync.WaitGroup{},
		doAfter:      make([]FuncWithStackInfo, 0),
		addStackInfo: debugBuild,
	}

	return waitGroup
}

type waitGroupParallel struct {
	lock              *sync.Mutex
	inner             *sync.WaitGroup
	errorGroupBuilder GroupBuilder
	doAfter           []FuncWithStackInfo

	addStackInfo bool

	isDone bool
}

func (waitGroup *waitGroupParallel) GetError() (err error) {
	waitGroup.wait()

	defer func() {
		if !waitGroup.errorGroupBuilder.Empty() {
			err = waitGroup.errorGroupBuilder.GetError()
		}
	}()

	for i := len(waitGroup.doAfter) - 1; i >= 0; i-- {
		doAfter := waitGroup.doAfter[i]
		err := doAfter.FuncErr()
		if err != nil {
			waitGroup.errorGroupBuilder.Add(doAfter.Wrap(err))
		}
	}

	return
}

func (waitGroup *waitGroupParallel) Do(f FuncErr) (d bool) {
	waitGroup.lock.Lock()

	if waitGroup.isDone {
		waitGroup.lock.Unlock()
		return false
	}

	waitGroup.lock.Unlock()

	waitGroup.inner.Add(1)

	var si stack_frame.Frame

	if waitGroup.addStackInfo {
		si, _ = stack_frame.MakeFrame(1)
	}

	go func() {
		err := f()

		waitGroup.doneWith(&si, err)
	}()

	return true
}

func (waitGroup *waitGroupParallel) DoAfter(f FuncErr) {
	waitGroup.lock.Lock()
	defer waitGroup.lock.Unlock()

	frame, _ := stack_frame.MakeFrame(1)

	waitGroup.doAfter = append(
		waitGroup.doAfter,
		FuncWithStackInfo{
			FuncErr: f,
			Frame:   frame,
		},
	)
}

func (waitGroup *waitGroupParallel) doneWith(
	frame *stack_frame.Frame,
	err error,
) {
	waitGroup.inner.Done()

	if err != nil {
		waitGroup.errorGroupBuilder.Add(frame.Wrap(err))
	}
}

func (waitGroup *waitGroupParallel) wait() {
	waitGroup.inner.Wait()

	waitGroup.lock.Lock()
	defer waitGroup.lock.Unlock()

	waitGroup.isDone = true
}
