package errors

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

func MakeWaitGroupParallel() WaitGroup {
	wg := &waitGroupParallel{
		lock:    &sync.Mutex{},
		inner:   &sync.WaitGroup{},
		err:     MakeMulti(),
		doAfter: make([]FuncWithStackInfo, 0),
	}

	return wg
}

type waitGroupParallel struct {
	lock    *sync.Mutex
	inner   *sync.WaitGroup
	err     Multi
	doAfter []FuncWithStackInfo

	addStackInfo bool

	isDone bool
}

func (wg *waitGroupParallel) GetError() (err error) {
	wg.wait()

	defer func() {
		if !wg.err.Empty() {
			err = wg.err
		}
	}()

	for i := len(wg.doAfter) - 1; i >= 0; i-- {
		doAfter := wg.doAfter[i]
		err := doAfter.FuncErr()
		if err != nil {
			wg.err.Add(doAfter.Wrap(err))
		}
	}

	return
}

func (wg *waitGroupParallel) Do(f FuncErr) (d bool) {
	wg.lock.Lock()

	if wg.isDone {
		wg.lock.Unlock()
		return false
	}

	wg.lock.Unlock()

	wg.inner.Add(1)

	var si stack_frame.Frame

	if wg.addStackInfo {
		si, _ = stack_frame.MakeFrame(1)
	}

	go func() {
		err := f()

		wg.doneWith(&si, err)
	}()

	return true
}

func (wg *waitGroupParallel) DoAfter(f FuncErr) {
	wg.lock.Lock()
	defer wg.lock.Unlock()

	frame, _ := stack_frame.MakeFrame(1)

	wg.doAfter = append(
		wg.doAfter,
		FuncWithStackInfo{
			FuncErr: f,
			Frame:   frame,
		},
	)
}

func (wg *waitGroupParallel) doneWith(frame *stack_frame.Frame, err error) {
	wg.inner.Done()

	if err != nil {
		wg.err.Add(frame.Wrap(err))
	}
}

func (wg *waitGroupParallel) wait() {
	wg.inner.Wait()

	wg.lock.Lock()
	defer wg.lock.Unlock()

	wg.isDone = true
}
