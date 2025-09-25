package stream_index

import (
	"fmt"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	State     int
	PageIndex = uint8
)

const (
	StateUnread = State(iota)
	StateChanged

	DigitWidth = 1
	PageCount  = 1 << (DigitWidth * 4)
)

type Index struct {
	hashType markl.FormatHash
	envRepo  env_repo.Env
	sunrise  ids.Tai
	preWrite interfaces.FuncIter[*sku.Transacted]
	path     string
	interfaces.NamedBlobAccess

	pages            [PageCount]writtenPage
	pagesAdded       [PageCount]writtenPage
	pagesAddedLatest [PageCount]writtenPage

	historicalChanges []string
	probeIndex
}

func MakeIndex(
	envRepo env_repo.Env,
	preWrite interfaces.FuncIter[*sku.Transacted],
	dir string,
	sunrise ids.Tai,
) (index *Index, err error) {
	index = &Index{
		hashType:        markl.FormatHashSha256,
		envRepo:         envRepo,
		sunrise:         sunrise,
		preWrite:        preWrite,
		path:            dir,
		NamedBlobAccess: envRepo,
	}

	if err = index.probeIndex.Initialize(
		envRepo,
		index.hashType,
	); err != nil {
		err = errors.Wrap(err)
		return index, err
	}

	if err = index.Initialize(); err != nil {
		err = errors.Wrap(err)
		return index, err
	}

	return index, err
}

func (index *Index) Initialize() (err error) {
	for n := range index.pages {
		index.pages[n].initialize(
			page_id.PageId{
				Prefix: "Page",
				Dir:    index.path,
				Index:  PageIndex(n),
			},
			index,
		)
	}

	return err
}

func (index *Index) GetPage(n PageIndex) (p *writtenPage) {
	p = &index.pages[n]
	return p
}

func (index *Index) GetProbeIndex() *probeIndex {
	return &index.probeIndex
}

func (index *Index) SetNeedsFlushHistory(changes []string) {
	index.historicalChanges = changes
}

func (index *Index) Flush(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	if len(index.historicalChanges) > 0 {
		if err = index.flushEverything(printerHeader); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else {
		if err = index.flushAdded(printerHeader); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (index *Index) flushAdded(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	ui.Log().Print("flushing")
	waitGroup := errors.MakeWaitGroupParallel()

	actualFlushCount := 0

	for n := range index.pages {
		if index.pages[n].additions.hasChanges() {
			ui.Log().Printf("actual flush for %d", n)
			actualFlushCount++
		}

		waitGroup.Do(index.makePageFlush(PageIndex(n), false))
	}

	if actualFlushCount > 0 {
		if err = printerHeader(
			fmt.Sprintf(
				"appending to index (%d/%d pages)",
				actualFlushCount,
				len(index.pages),
			),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	waitGroup.DoAfter(index.index.Flush)

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if actualFlushCount > 0 {
		if err = printerHeader(
			fmt.Sprintf(
				"appended to index (%d/%d pages)",
				actualFlushCount,
				len(index.pages),
			),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (index *Index) flushEverything(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	ui.Log().Print("flushing")
	waitGroup := errors.MakeWaitGroupParallel()

	for n := range index.pages {
		waitGroup.Do(index.makePageFlush(PageIndex(n), true))
	}

	if err = printerHeader(
		fmt.Sprintf(
			"writing index (%d pages)",
			len(index.pages),
		),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for n, change := range index.historicalChanges {
		if err = printerHeader(fmt.Sprintf("change: %s", change)); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if n == 99 {
			if err = printerHeader(
				fmt.Sprintf(
					"(%d more changes omitted)",
					len(index.historicalChanges)-100,
				),
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

			break
		}
	}

	waitGroup.DoAfter(index.index.Flush)

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = printerHeader(
		fmt.Sprintf(
			"wrote index (%d pages)",
			len(index.pages),
		),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func PageIndexForObject(
	width PageIndex,
	object *sku.Transacted,
	hashType interfaces.FormatHash,
) (n PageIndex, err error) {
	if n, err = PageIndexForObjectId(
		width,
		object.GetObjectId(),
		hashType,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func PageIndexForObjectId(
	width PageIndex,
	oid *ids.ObjectId,
	hashType interfaces.FormatHash,
) (n PageIndex, err error) {
	if n, err = page_id.PageIndexForString(
		width,
		oid.String(),
		hashType,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (index *Index) Add(
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	var pageIndex PageIndex

	if pageIndex, err = PageIndexForObject(
		DigitWidth,
		object,
		index.hashType,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = index.add(pageIndex, object, options); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO add support for *errors.Context closure
func (index *Index) ReadPrimitiveQuery(
	queryGroup sku.PrimitiveQueryGroup,
	funcIter interfaces.FuncIter[*sku.Transacted],
) (err error) {
	ui.Log().Printf("starting query: %q", queryGroup)
	waitGroup := &sync.WaitGroup{}
	ch := make(chan struct{}, PageCount)
	groupBuilder := errors.MakeGroupBuilder()
	chDone := make(chan struct{})

	isDone := func() bool {
		select {
		case <-chDone:
			return true

		default:
			return false
		}
	}

	funcIter = pool.MakePooledChain(
		sku.GetTransactedPool(),
		funcIter,
	)

	// TODO switch to errors.MakeWaitGroupParallel()
	for n := range index.pages {
		waitGroup.Add(1)

		go func(pageIndex PageIndex, openFileCh chan struct{}) {
			pageReader, pageReaderClose := index.makePageReader(pageIndex)
			defer errors.Deferred(&err, pageReaderClose)

			ui.Log().Printf(
				"starting query on page %d: %q",
				pageReader.pageId.Index,
				queryGroup,
			)
			defer waitGroup.Done()
			defer func() {
				openFileCh <- struct{}{}
			}()

			for !isDone() {
				var err1 error

				if err1 = pageReader.readFull(
					queryGroup,
					funcIter,
					pageReadOptions{
						includeAdded:       false,
						includeAddedLatest: false,
					},
				); err1 != nil {
					if isDone() {
						break
					}

					switch {
					case errors.IsTooManyOpenFiles(err1):
						<-openFileCh
						continue

					case errors.IsStopIteration(err1):

					default:
						groupBuilder.Add(err1)
					}
				}

				break
			}
		}(PageIndex(n), ch)
	}

	waitGroup.Wait()

	if groupBuilder.Len() > 0 {
		err = groupBuilder.GetError()
	}

	return err
}
