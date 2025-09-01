package stream_index

import (
	"fmt"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type State int

const (
	StateUnread = State(iota)
	StateChanged

	DigitWidth = 1
	PageCount  = 1 << (DigitWidth * 4)
)

type Index struct {
	hashType markl.HashType
	envRepo  env_repo.Env
	sunrise  ids.Tai
	preWrite interfaces.FuncIter[*sku.Transacted]
	path     string
	interfaces.CacheIOFactory
	pages             [PageCount]Page
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
		hashType:       markl.HashTypeSha256,
		envRepo:        envRepo,
		sunrise:        sunrise,
		preWrite:       preWrite,
		path:           dir,
		CacheIOFactory: envRepo,
	}

	if err = index.probeIndex.Initialize(
		envRepo,
		index.hashType,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = index.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (index *Index) Initialize() (err error) {
	for n := range index.pages {
		index.pages[n].initialize(
			page_id.PageId{
				Prefix: "Page",
				Dir:    index.path,
				Index:  uint8(n),
			},
			index,
		)
	}

	return
}

func (index *Index) GetPage(n uint8) (p *Page) {
	p = &index.pages[n]
	return
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
			return
		}
	} else {
		if err = index.flushAdded(printerHeader); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (index *Index) flushAdded(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	ui.Log().Print("flushing")
	wg := errors.MakeWaitGroupParallel()

	actualFlushCount := 0

	for n := range index.pages {
		if index.pages[n].hasChanges {
			ui.Log().Printf("actual flush for %d", n)
			actualFlushCount++
		}

		wg.Do(index.pages[n].MakeFlush(false))
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
			return
		}
	}

	wg.DoAfter(index.Index.Flush)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
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
			return
		}
	}

	return
}

func (index *Index) flushEverything(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	ui.Log().Print("flushing")
	wg := errors.MakeWaitGroupParallel()

	for n := range index.pages {
		wg.Do(index.pages[n].MakeFlush(true))
	}

	if err = printerHeader(
		fmt.Sprintf(
			"writing index (%d pages)",
			len(index.pages),
		),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	for n, change := range index.historicalChanges {
		if err = printerHeader(fmt.Sprintf("change: %s", change)); err != nil {
			err = errors.Wrap(err)
			return
		}

		if n == 99 {
			if err = printerHeader(
				fmt.Sprintf(
					"(%d more changes omitted)",
					len(index.historicalChanges)-100,
				),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			break
		}
	}

	wg.DoAfter(index.Index.Flush)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = printerHeader(
		fmt.Sprintf(
			"wrote index (%d pages)",
			len(index.pages),
		),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func PageIndexForObject(
	width uint8,
	object *sku.Transacted,
	hashType interfaces.HashType,
) (n uint8, err error) {
	if n, err = PageIndexForObjectId(
		width,
		object.GetObjectId(),
		hashType,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func PageIndexForObjectId(
	width uint8,
	oid *ids.ObjectId,
	hashType interfaces.HashType,
) (n uint8, err error) {
	if n, err = page_id.PageIndexForString(
		width,
		oid.String(),
		hashType,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (index *Index) Add(
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	var n uint8

	if n, err = PageIndexForObject(
		DigitWidth,
		object,
		index.hashType,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	p := index.GetPage(n)

	if err = p.add(object, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO rename
func (index *Index) ReadOneSha(
	blobId interfaces.MarklId,
	object *sku.Transacted,
) (err error) {
	errors.PanicIfError(markl.AssertIdIsNotNull(blobId, "index lookup"))

	var loc object_probe_index.Loc

	if loc, err = index.readOneMarklIdLoc(blobId); err != nil {
		return
	}

	if err = index.readOneLoc(loc, object); err != nil {
		return
	}

	return
}

// TODO rename
func (index *Index) ReadManySha(
	blobId interfaces.MarklId,
) (objects []*sku.Transacted, err error) {
	var locs []object_probe_index.Loc

	if locs, err = index.readManyMarklIdLoc(blobId); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, loc := range locs {
		sk := sku.GetTransactedPool().Get()

		if err = index.readOneLoc(loc, sk); err != nil {
			err = errors.Wrapf(err, "Loc: %s", loc)
			return
		}

		objects = append(objects, sk)
	}

	return
}

func (index *Index) ObjectExists(
	objectId *ids.ObjectId,
) (err error) {
	var n uint8

	objectIdString := objectId.String()

	if n, err = page_id.PageIndexForString(
		DigitWidth,
		objectIdString,
		index.hashType,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	p := index.GetPage(n)

	if _, ok := p.addedObjectIdLookup[objectIdString]; ok {
		return
	}

	digest := markl.HashTypeSha256.FromStringContent(objectIdString)
	defer markl.PutBlobId(digest)

	if _, err = index.readOneMarklIdLoc(digest); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (index *Index) ReadOneObjectId(
	objectId interfaces.ObjectId,
	object *sku.Transacted,
) (err error) {
	objectIdString := objectId.String()

	if objectIdString == "" {
		panic("empty object id")
	}

	digest, repool := markl.HashTypeSha256.GetMarklIdForString(
		objectIdString,
	)
	defer repool()

	if err = index.ReadOneSha(digest, object); err != nil {
		return
	}

	return
}

func (index *Index) ReadManyObjectId(
	id interfaces.ObjectId,
) (skus []*sku.Transacted, err error) {
	sh := markl.HashTypeSha256.FromStringContent(id.String())
	defer markl.PutBlobId(sh)

	if skus, err = index.ReadManySha(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO switch to empty=not found semantics instead of error
func (index *Index) ReadOneObjectIdTai(
	k interfaces.ObjectId,
	t ids.Tai,
) (sk *sku.Transacted, err error) {
	if t.IsEmpty() {
		err = collections.MakeErrNotFoundString(t.String())
		return
	}

	sh := markl.HashTypeSha256.FromStringContent(k.String() + t.String())
	defer markl.PutBlobId(sh)

	sk = sku.GetTransactedPool().Get()

	if err = index.ReadOneSha(sh, sk); err != nil {
		return
	}

	return
}

func (index *Index) readOneLoc(
	loc object_probe_index.Loc,
	sk *sku.Transacted,
) (err error) {
	p := index.pages[loc.Page]

	if err = p.readOneRange(loc.Range, sk); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO add support for *errors.Context closure
func (index *Index) ReadPrimitiveQuery(
	qg sku.PrimitiveQueryGroup,
	funcIter interfaces.FuncIter[*sku.Transacted],
) (err error) {
	ui.Log().Printf("starting query: %q", qg)
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

	for n := range index.pages {
		waitGroup.Add(1)

		go func(p *Page, openFileCh chan struct{}) {
			ui.Log().Printf("starting query on page %d: %q", p.PageId.Index, qg)
			defer waitGroup.Done()
			defer func() {
				openFileCh <- struct{}{}
			}()

			for !isDone() {
				var err1 error

				if err1 = p.copyHistoryAndMaybeLatest(
					qg,
					funcIter,
					false,
					false,
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
		}(&index.pages[n], ch)
	}

	waitGroup.Wait()

	if groupBuilder.Len() > 0 {
		err = groupBuilder.GetError()
	}

	return
}
