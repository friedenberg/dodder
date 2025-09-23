package object_probe_index

import (
	"io"

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

func (index *Index) AddDigest(
	digest interfaces.MarklId,
	loc Loc,
) (err error) {
	return index.addDigest(digest, loc)
}

func (index *Index) addDigest(
	digest interfaces.MarklId,
	loc Loc,
) (err error) {
	if digest.IsNull() {
		return err
	}

	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		digest,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if digest.GetMarklFormat().GetMarklFormatId() != index.hashType.GetMarklFormatId() {
		replacementId, repool := index.hashType.GetMarklIdForMarklId(digest)
		defer repool()

		// ui.Debug().Print(
		// 	digest.StringWithFormat(),
		// 	"->",
		// 	replacementId.StringWithFormat(),
		// )

		digest = replacementId
	}

	if err = index.pages[pageIndex].AddMarklId(digest, loc); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return err
	}

	return err
}

func (index *Index) ReadOne(
	digest interfaces.MarklId,
) (loc Loc, err error) {
	actual := digest.GetMarklFormat().GetMarklFormatId()

	if actual != index.hashType.GetMarklFormatId() {
		err = errors.Errorf("unsupported hash type: %q", actual)
		return loc, err
	}

	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		digest,
	); err != nil {
		err = errors.Wrap(err)
		return loc, err
	}

	return index.pages[pageIndex].ReadOne(digest)
}

func (index *Index) ReadMany(
	digest interfaces.MarklId,
	locations *[]Loc,
) (err error) {
	actual := digest.GetMarklFormat().GetMarklFormatId()

	if actual != index.hashType.GetMarklFormatId() {
		err = errors.Errorf("unsupported hash type: %q", actual)
		return err
	}

	var pageIndex uint8

	if pageIndex, err = page_id.PageIndexForDigest(
		DigitWidth,
		digest,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return index.pages[pageIndex].ReadMany(digest, locations)
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
