package debug

import (
	"bufio"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

type Context struct {
	bufferedWriterTrace                    *bufio.Writer
	filePprofCpu, filePprofHeap, fileTrace *os.File
	options                                Options
}

func MakeContext(
	ctx interfaces.Context,
	options Options,
) (c *Context, err error) {
	c = &Context{
		options: options,
	}

	if options.ExitOnMemoryExhaustion {
		// TODO start memory limit struct instead
		ticker := time.NewTicker(time.Millisecond)
		ctx.After(
			errors.MakeFuncContextFromFuncNil(ticker.Stop),
		)

		var cgroupMemoryLimit uint64

		if cgroupMemoryLimit, err = getMemoryLimit(); err != nil {
			cgroupMemoryLimit = 1500 * 1024 * 1024 // 1.5 GB
			ui.Err().Printf(
				"memory limit not found, setting to %s",
				ui.GetHumanBytesString(cgroupMemoryLimit),
			)

			err = nil
			// err = errors.Wrapf(err, "could not determine memory limit")
			// return
		}

		var memOnce sync.Once

		go func() {
			var memStats runtime.MemStats

			for {
				select {
				case <-ctx.Done():
					return

				case <-ticker.C:
					runtime.ReadMemStats(&memStats)
					memoryInUse := memStats.Alloc

					percent := float64(
						memoryInUse,
					) / float64(
						cgroupMemoryLimit,
					) * 100

					if percent >= 90 {
						memOnce.Do(
							func() {
								ui.Err().Printf(
									"%.2f%% memory used: %s of %s",
									percent,
									ui.GetHumanBytesString(memoryInUse),
									ui.GetHumanBytesString(cgroupMemoryLimit),
								)

								func() {
									defer func() {
										recover()
									}()

									errors.ContextCancelWithErrorf(
										ctx,
										"10%% memory remaining",
									)
								}()
							},
						)
					}
				}
			}
		}()
	}

	if options.GCDisabled {
		debug.SetGCPercent(-1)
	}

	if options.PProfCPU {
		if c.filePprofCpu, err = files.Create("cpu.pprof"); err != nil {
			err = errors.Wrap(err)
			return
		}

		pprof.StartCPUProfile(c.filePprofCpu)
	}

	if options.PProfHeap {
		if c.filePprofHeap, err = files.Create("heap.pprof"); err != nil {
			err = errors.Wrap(err)
			return
		}

		pprof.StartCPUProfile(c.filePprofCpu)
	}

	if options.Trace {
		if c.fileTrace, err = files.Create("trace"); err != nil {
			err = errors.Wrap(err)
			return
		}

		c.bufferedWriterTrace = bufio.NewWriter(c.fileTrace)
		trace.Start(c.bufferedWriterTrace)
	}

	if options.GCDisabled {
		debug.SetGCPercent(-1)
	}

	ctx.After(errors.MakeFuncContextFromFuncErr(c.Close))

	return
}

func (c *Context) Close() error {
	waitGroupStopOrWrite := errors.MakeWaitGroupParallel()
	multiError := errors.MakeMulti()

	if c.fileTrace != nil {
		waitGroupStopOrWrite.Do(errors.MakeFuncErrFromFuncNil(trace.Stop))
	}

	if c.filePprofCpu != nil {
		waitGroupStopOrWrite.Do(
			errors.MakeFuncErrFromFuncNil(pprof.StopCPUProfile),
		)
	}

	if c.filePprofHeap != nil {
		waitGroupStopOrWrite.Do(func() error {
			return pprof.Lookup("heap").WriteTo(c.filePprofHeap, 0)
		})
	}

	if err := waitGroupStopOrWrite.GetError(); err != nil {
		multiError.Add(errors.Wrap(err))
	}

	if c.fileTrace != nil {
		if err := c.bufferedWriterTrace.Flush(); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	waitGroupClose := errors.MakeWaitGroupParallel()

	if c.fileTrace != nil {
		waitGroupClose.Do(c.fileTrace.Close)
	}

	if c.filePprofCpu != nil {
		waitGroupClose.Do(c.filePprofCpu.Close)
	}

	if c.options.PProfHeap {
		waitGroupClose.Do(c.filePprofHeap.Close)
	}

	if err := waitGroupClose.GetError(); err != nil {
		multiError.Add(errors.Wrap(err))
	}

	if multiError.Len() > 0 {
		return multiError
	}

	return nil
}
