package object_probe_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
)

type (
	// TODO add support for different digests
	commonInterface interface {
		// TODO rename to AddDigest and enforce digest type
		AddMarklId(interfaces.MarklId, Loc) error
		ReadOne(id interfaces.MarklId) (loc Loc, err error)
		ReadMany(id interfaces.MarklId, locs *[]Loc) (err error)
	}

	Index interface {
		commonInterface
		PrintAll(env_ui.Env) error
		errors.Flusher
	}
)

type Metadata = object_metadata.Metadata

const (
	DigitWidth = 1
	PageCount  = 1 << (DigitWidth * 4)
)

type index struct {
	hashType markl.HashType
	rowWidth int
	pages    [PageCount]page
}

func MakePermitDuplicates(
	envRepo env_repo.Env,
	path string,
	hashType markl.HashType,
) (indecks *index, err error) {
	indecks = &index{hashType: hashType}
	err = indecks.initialize(rowEqualerComplete{}, envRepo, path)
	return
}

func MakeNoDuplicates(
	envRepo env_repo.Env,
	dir string,
	hashType markl.HashType,
) (indecks *index, err error) {
	indecks = &index{hashType: hashType}
	err = indecks.initialize(rowEqualerShaOnly{}, envRepo, dir)
	return
}

func (index *index) initialize(
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

func (index *index) GetObjectProbeIndex() Index {
	return index
}

func (index *index) AddMarklId(
	blobId interfaces.MarklId,
	loc Loc,
) (err error) {
	return index.addBlobId(blobId, loc)
}

func (index *index) addBlobId(
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

func (index *index) ReadOne(
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

func (index *index) ReadMany(
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

func (index *index) ReadOneKey(
	formatKey string,
	metadata *object_metadata.Metadata,
) (loc Loc, err error) {
	var format object_inventory_format.Format

	if format, err = object_inventory_format.FormatForMarklFormatIdError(formatKey); err != nil {
		err = errors.Wrap(err)
		return
	}

	var blobId interfaces.MarklId

	if blobId, err = object_inventory_format.GetDigestForMetadata(
		format,
		metadata,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer markl.PutBlobId(blobId)

	if loc, err = index.ReadOne(blobId); err != nil {
		err = errors.Wrapf(err, "Key: %s", formatKey)
		return
	}

	return
}

// TODO remove formatKey and switch to setting formatter on init
func (index *index) ReadManyKeys(
	formatKey string,
	metadata *object_metadata.Metadata,
	locations *[]Loc,
) (err error) {
	var format object_inventory_format.Format

	if format, err = object_inventory_format.FormatForMarklFormatIdError(
		formatKey,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var blobId interfaces.MarklId

	if blobId, err = object_inventory_format.GetDigestForMetadata(
		format,
		metadata,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.ReadMany(blobId, locations)
}

func (index *index) ReadAll(
	metadata *object_metadata.Metadata,
	locations *[]Loc,
) (err error) {
	var blobIds map[string]interfaces.MarklId

	if blobIds, err = object_inventory_format.GetDigestsForMetadata(
		metadata,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	wg := errors.MakeWaitGroupParallel()

	for key, blobId := range blobIds {
		wg.Do(
			func() (err error) {
				var loc Loc

				if loc, err = index.ReadOne(blobId); err != nil {
					err = errors.Wrapf(err, "Key: %s", key)
					return
				}

				*locations = append(*locations, loc)

				return
			},
		)
	}

	return wg.GetError()
}

func (index *index) PrintAll(env env_ui.Env) (err error) {
	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]

		if err = page.PrintAll(env); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (index *index) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()

	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]
		wg.Do(page.Flush)
	}

	return wg.GetError()
}
