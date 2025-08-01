package stream_index

import (
	"fmt"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type State int

const (
	StateUnread = State(iota)
	StateChanged
)

type PageDelegate interface {
	ShouldAddVerzeichnisse(*sku.Transacted) error
	ShouldFlushVerzeichnisse(*sku.Transacted) error
}

type PageDelegateGetter interface {
	GetVerzeichnissePageDelegate(uint8) PageDelegate
}

const (
	DigitWidth = 1
	PageCount  = 1 << (DigitWidth * 4)
)

var options object_inventory_format.Options

func init() {
	options = object_inventory_format.Options{
		Tai:           true,
		Verzeichnisse: true,
		PrintFinalSha: true,
	}
}

type Index struct {
	directoryLayout env_repo.Env
	sunrise         ids.Tai
	preWrite        interfaces.FuncIter[*sku.Transacted]
	path            string
	interfaces.CacheIOFactory
	pages             [PageCount]Page
	historicalChanges []string
	probe_index
}

func MakeIndex(
	s env_repo.Env,
	preWrite interfaces.FuncIter[*sku.Transacted],
	dir string,
	sunrise ids.Tai,
) (i *Index, err error) {
	i = &Index{
		directoryLayout: s,
		sunrise:         sunrise,
		preWrite:        preWrite,
		path:            dir,
		CacheIOFactory:  s,
	}

	if err = i.probe_index.Initialize(
		s,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = i.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (i *Index) Initialize() (err error) {
	for n := range i.pages {
		i.pages[n].initialize(
			page_id.PageId{
				Prefix: "Page",
				Dir:    i.path,
				Index:  uint8(n),
			},
			i,
		)
	}

	return
}

func (i *Index) GetPage(n uint8) (p *Page) {
	p = &i.pages[n]
	return
}

func (i *Index) GetProbeIndex() *probe_index {
	return &i.probe_index
}

func (i *Index) SetNeedsFlushHistory(changes []string) {
	i.historicalChanges = changes
}

func (i *Index) Flush(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	if len(i.historicalChanges) > 0 {
		if err = i.flushEverything(printerHeader); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = i.flushAdded(printerHeader); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (i *Index) flushAdded(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	ui.Log().Print("flushing")
	wg := errors.MakeWaitGroupParallel()

	actualFlushCount := 0

	for n := range i.pages {
		if i.pages[n].hasChanges {
			ui.Log().Printf("actual flush for %d", n)
			actualFlushCount++
		}

		wg.Do(i.pages[n].MakeFlush(false))
	}

	if actualFlushCount > 0 {
		if err = printerHeader(
			fmt.Sprintf(
				"appending to index (%d/%d pages)",
				actualFlushCount,
				len(i.pages),
			),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	wg.DoAfter(i.Index.Flush)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if actualFlushCount > 0 {
		if err = printerHeader(
			fmt.Sprintf(
				"appended to index (%d/%d pages)",
				actualFlushCount,
				len(i.pages),
			),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (i *Index) flushEverything(
	printerHeader interfaces.FuncIter[string],
) (err error) {
	ui.Log().Print("flushing")
	wg := errors.MakeWaitGroupParallel()

	for n := range i.pages {
		wg.Do(i.pages[n].MakeFlush(true))
	}

	if err = printerHeader(
		fmt.Sprintf(
			"writing index (%d pages)",
			len(i.pages),
		),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	for n, change := range i.historicalChanges {
		if err = printerHeader(fmt.Sprintf("change: %s", change)); err != nil {
			err = errors.Wrap(err)
			return
		}

		if n == 99 {
			if err = printerHeader(
				fmt.Sprintf(
					"(%d more changes omitted)",
					len(i.historicalChanges)-100,
				),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			break
		}
	}

	wg.DoAfter(i.Index.Flush)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = printerHeader(
		fmt.Sprintf(
			"wrote index (%d pages)",
			len(i.pages),
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
) (n uint8, err error) {
	if n, err = PageIndexForObjectId(width, object.GetObjectId()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func PageIndexForObjectId(width uint8, oid *ids.ObjectId) (n uint8, err error) {
	if n, err = page_id.PageIndexForString(
		width,
		oid.String(),
		sha.Env{},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (i *Index) Add(
	z *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	var n uint8

	if n, err = PageIndexForObject(
		DigitWidth,
		z,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	p := i.GetPage(n)

	if err = p.add(z, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Index) ReadOneSha(
	sh interfaces.BlobId,
	sk *sku.Transacted,
) (err error) {
	var loc object_probe_index.Loc

	if loc, err = s.readOneShaLoc(sh); err != nil {
		return
	}

	if err = s.readOneLoc(loc, sk); err != nil {
		return
	}

	return
}

func (s *Index) ReadManySha(
	sh interfaces.BlobId,
) (skus []*sku.Transacted, err error) {
	var locs []object_probe_index.Loc

	if locs, err = s.readManyShaLoc(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, loc := range locs {
		sk := sku.GetTransactedPool().Get()

		if err = s.readOneLoc(loc, sk); err != nil {
			err = errors.Wrapf(err, "Loc: %s", loc)
			return
		}

		skus = append(skus, sk)
	}

	return
}

func (s *Index) ObjectExists(
	objectId *ids.ObjectId,
) (err error) {
	var n uint8

	objectIdString := objectId.String()

	if n, err = page_id.PageIndexForString(
		DigitWidth,
		objectIdString,
		sha.Env{},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	p := s.GetPage(n)

	if _, ok := p.oids[objectIdString]; ok {
		return
	}

	sh := sha.FromStringContent(objectIdString)
	defer digests.PutBlobId(sh)

	if _, err = s.readOneShaLoc(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Index) ReadOneObjectId(
	oid interfaces.ObjectId,
	sk *sku.Transacted,
) (err error) {
	sh := sha.FromStringContent(oid.String())
	defer digests.PutBlobId(sh)

	if err = s.ReadOneSha(sh, sk); err != nil {
		return
	}

	return
}

func (s *Index) ReadManyObjectId(
	id interfaces.ObjectId,
) (skus []*sku.Transacted, err error) {
	sh := sha.FromStringContent(id.String())
	defer digests.PutBlobId(sh)

	if skus, err = s.ReadManySha(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO switch to empty=not found semantics instead of error
func (s *Index) ReadOneObjectIdTai(
	k interfaces.ObjectId,
	t ids.Tai,
) (sk *sku.Transacted, err error) {
	if t.IsEmpty() {
		err = collections.MakeErrNotFoundString(t.String())
		return
	}

	sh := sha.FromStringContent(k.String() + t.String())
	defer digests.PutBlobId(sh)

	sk = sku.GetTransactedPool().Get()

	if err = s.ReadOneSha(sh, sk); err != nil {
		return
	}

	return
}

func (s *Index) readOneLoc(
	loc object_probe_index.Loc,
	sk *sku.Transacted,
) (err error) {
	p := s.pages[loc.Page]

	if err = p.readOneRange(loc.Range, sk); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO add support for *errors.Context closure
func (i *Index) ReadPrimitiveQuery(
	qg sku.PrimitiveQueryGroup,
	funcIter interfaces.FuncIter[*sku.Transacted],
) (err error) {
	ui.Log().Printf("starting query: %q", qg)
	waitGroup := &sync.WaitGroup{}
	ch := make(chan struct{}, PageCount)
	multiError := errors.MakeGroupBuilder()
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

	for n := range i.pages {
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
						multiError.Add(err1)
					}
				}

				break
			}
		}(&i.pages[n], ch)
	}

	waitGroup.Wait()

	if multiError.Len() > 0 {
		err = multiError
	}

	return
}
