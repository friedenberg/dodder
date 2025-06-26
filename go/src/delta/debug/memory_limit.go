package debug

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

type memoryLimit struct {
	ctx errors.Context
	*time.Ticker
	memoryLimit atomic.Uint64
	memoryInUse atomic.Uint64
	sync.Once
}

func (ml *memoryLimit) Start(ctx errors.Context) (err error) {
	ml.ctx = ctx

	ml.Ticker = time.NewTicker(time.Millisecond)
	ctx.After(errors.MakeNilFunc(ml.Stop))

	var memoryLimit uint64

	if memoryLimit, err = getMemoryLimit(); err != nil {
		memoryLimit = (1500 * 1024 * 1024) // 1.5 GB
		ui.Err().Printf(
			"memory limit not found, setting to %s",
			ui.GetHumanBytesString(memoryLimit),
		)

		err = nil
		// err = errors.Wrapf(err, "could not determine memory limit")
		// return
	}

	ml.memoryLimit.Swap(memoryLimit)

	go func() {
		var memStats runtime.MemStats

		for {
			select {
			case <-ctx.Done():
				return

			case <-ml.C:
				runtime.ReadMemStats(&memStats)
				ml.memoryInUse.Swap(memStats.Alloc)

				memoryInUse := ml.memoryInUse.Load()
				memoryLimit := ml.memoryLimit.Load()

				percent := float64(memoryInUse) / float64(memoryLimit) * 100

				if percent >= 90 {
					ml.Do(ml.Terminate)
				}
			}
		}
	}()

	return
}

func (ml *memoryLimit) Terminate() {
	memoryInUse := ml.memoryInUse.Load()
	memoryLimit := ml.memoryLimit.Load()
	percent := float64(memoryInUse) / float64(memoryLimit) * 100

	ui.Err().Printf(
		"%.2f%% memory used: %s of %s",
		percent,
		ui.GetHumanBytesString(memoryInUse),
		ui.GetHumanBytesString(memoryLimit),
	)

	defer func() {
		recover()
	}()

	ml.ctx.CancelWithErrorf("10% memory remaining")
}
