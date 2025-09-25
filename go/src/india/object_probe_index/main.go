package object_probe_index

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

const (
	DigitWidth = 1
	PageCount  = 1 << (DigitWidth * 4)
)

type Index struct {
	hashType markl.FormatHash
	rowSize  int
	pages    [PageCount]page
}

func MakePermitDuplicates(
	envRepo env_repo.Env,
	path string,
	hashType markl.FormatHash,
) (indecks *Index, err error) {
	indecks = &Index{hashType: hashType}
	err = indecks.initialize(rowEqualerComplete{}, envRepo, path)
	return indecks, err
}

func MakeNoDuplicates(
	envRepo env_repo.Env,
	dir string,
	hashType markl.FormatHash,
) (indecks *Index, err error) {
	indecks = &Index{hashType: hashType}
	err = indecks.initialize(rowEqualerDigestOnly{}, envRepo, dir)
	return indecks, err
}

func (index *Index) initialize(
	equaler interfaces.Equaler[*row],
	envRepo env_repo.Env,
	dir string,
) (err error) {
	index.rowSize = index.hashType.GetSize() + 1 + 8 + 8

	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]
		page.initialize(
			equaler,
			envRepo,
			page_id.PageIdFromPath(uint8(pageIndex), dir),
			index.hashType,
			index.rowSize,
		)
	}

	return err
}

func (index *Index) GetHashType() markl.FormatHash {
	return index.hashType
}

func (index *Index) AddDigest(
	probeId ids.ProbeIdWithObjectId,
	loc Loc,
) (err error) {
	if probeId.Id.IsNull() {
		return err
	}

	id := probeId.Id

	if probeId.Id.GetMarklFormat().GetMarklFormatId() != index.hashType.GetMarklFormatId() {
		replacementId, repool := index.hashType.GetMarklIdForMarklId(probeId.Id)
		defer repool()

		id = replacementId
	}

	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		id,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = index.pages[pageIndex].AddMarklId(id, loc); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return err
	}

	return err
}

func (index *Index) ReadOne(
	originalId interfaces.MarklId,
) (loc Loc, err error) {
	id := originalId

	if id.GetMarklFormat().GetMarklFormatId() != index.hashType.GetMarklFormatId() {
		replacementId, repool := index.hashType.GetMarklIdForMarklId(id)
		defer repool()

		id = replacementId
	}

	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		id,
	); err != nil {
		return loc, err
	}

	if loc, err = index.pages[pageIndex].ReadOne(id); err != nil {
		return loc, err
	}

	return loc, err
}

func (index *Index) ReadMany(
	originalId interfaces.MarklId,
	locations *[]Loc,
) (err error) {
	id := originalId

	if id.GetMarklFormat().GetMarklFormatId() != index.hashType.GetMarklFormatId() {
		replacementId, repool := index.hashType.GetMarklIdForMarklId(id)
		defer repool()

		id = replacementId
	}

	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		id,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = index.pages[pageIndex].ReadMany(id, locations); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (index *Index) PrintAll(env env_ui.Env) (err error) {
	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]

		if err = page.PrintAll(env); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (index *Index) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()

	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]
		wg.Do(page.Flush)
	}

	return wg.GetError()
}
