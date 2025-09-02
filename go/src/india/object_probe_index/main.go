package object_probe_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

const (
	DigitWidth = 1
	PageCount  = 1 << (DigitWidth * 4)
)

type Index struct {
	hashType markl.HashType
	rowWidth int
	pages    [PageCount]page
}

func MakePermitDuplicates(
	envRepo env_repo.Env,
	path string,
	hashType markl.HashType,
) (indecks *Index, err error) {
	indecks = &Index{hashType: hashType}
	err = indecks.initialize(rowEqualerComplete{}, envRepo, path)
	return
}

func MakeNoDuplicates(
	envRepo env_repo.Env,
	dir string,
	hashType markl.HashType,
) (indecks *Index, err error) {
	indecks = &Index{hashType: hashType}
	err = indecks.initialize(rowEqualerShaOnly{}, envRepo, dir)
	return
}

func (index *Index) initialize(
	equaler interfaces.Equaler[*row],
	envRepo env_repo.Env,
	dir string,
) (err error) {
	index.rowWidth = index.hashType.GetSize() + 1 + 8 + 8

	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]
		page.initialize(
			equaler,
			envRepo,
			page_id.PageIdFromPath(uint8(pageIndex), dir),
			index.hashType,
			index.rowWidth,
		)
	}

	return
}

func (index *Index) AddMarklId(
	blobId interfaces.MarklId,
	loc Loc,
) (err error) {
	return index.addBlobId(blobId, loc)
}

func (index *Index) addBlobId(
	blobId interfaces.MarklId,
	loc Loc,
) (err error) {
	if blobId.IsNull() {
		return
	}

	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		blobId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.pages[pageIndex].AddMarklId(blobId, loc)
}

func (index *Index) ReadOne(
	blobId interfaces.MarklId,
) (loc Loc, err error) {
	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		blobId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.pages[pageIndex].ReadOne(blobId)
}

func (index *Index) ReadMany(
	blobId interfaces.MarklId,
	locations *[]Loc,
) (err error) {
	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		blobId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.pages[pageIndex].ReadMany(blobId, locations)
}

func (index *Index) PrintAll(env env_ui.Env) (err error) {
	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]

		if err = page.PrintAll(env); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (index *Index) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()

	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]
		wg.Do(page.Flush)
	}

	return wg.GetError()
}
