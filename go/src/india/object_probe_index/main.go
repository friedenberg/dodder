package object_probe_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
)

type (
	// TODO add support for different digests
	commonInterface interface {
		// TODO rename to AddDigest and enforce digest type
		AddBlobId(interfaces.BlobId, Loc) error
		ReadOne(id interfaces.BlobId) (loc Loc, err error)
		ReadMany(id interfaces.BlobId, locs *[]Loc) (err error)
	}

	pageInterface interface {
		GetObjectProbeIndexPage() pageInterface
		commonInterface
		PrintAll(env_ui.Env) error
		errors.Flusher
	}

	Index interface {
		GetObjectProbeIndex() Index
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

type object_probe_index struct {
	hashType merkle.HashType
	pages    [PageCount]page
}

func MakePermitDuplicates(
	envRepo env_repo.Env,
	path string,
	hashType merkle.HashType,
) (index *object_probe_index, err error) {
	index = &object_probe_index{hashType: hashType}
	err = index.initialize(rowEqualerComplete{}, envRepo, path)
	return
}

func MakeNoDuplicates(
	envRepo env_repo.Env,
	dir string,
	hashType merkle.HashType,
) (index *object_probe_index, err error) {
	index = &object_probe_index{hashType: hashType}
	err = index.initialize(rowEqualerShaOnly{}, envRepo, dir)
	return
}

func (index *object_probe_index) initialize(
	equaler interfaces.Equaler[*row],
	envRepo env_repo.Env,
	dir string,
) (err error) {
	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]
		page.initialize(
			equaler,
			envRepo,
			page_id.PageIdFromPath(uint8(pageIndex), dir),
			index.hashType,
		)
	}

	return
}

func (index *object_probe_index) GetObjectProbeIndex() Index {
	return index
}

func (index *object_probe_index) AddMetadata(m *Metadata, loc Loc) (err error) {
	var blobIds map[string]interfaces.BlobId

	if blobIds, err = object_inventory_format.GetDigestsForMetadata(m); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, blobId := range blobIds {
		if err = index.addBlobId(blobId, loc); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (index *object_probe_index) AddBlobId(
	blobId interfaces.BlobId,
	loc Loc,
) (err error) {
	return index.addBlobId(blobId, loc)
}

func (index *object_probe_index) addBlobId(
	blobId interfaces.BlobId,
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

	return index.pages[pageIndex].AddBlobId(blobId, loc)
}

func (index *object_probe_index) ReadOne(
	blobId interfaces.BlobId,
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

func (index *object_probe_index) ReadMany(
	blobId interfaces.BlobId,
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

func (index *object_probe_index) ReadOneKey(
	formatKey string,
	metadata *object_metadata.Metadata,
) (loc Loc, err error) {
	var format object_inventory_format.Format

	if format, err = object_inventory_format.FormatForKeyError(formatKey); err != nil {
		err = errors.Wrap(err)
		return
	}

	var blobId interfaces.BlobId

	if blobId, err = object_inventory_format.GetDigestForMetadata(
		format,
		metadata,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer merkle_ids.PutBlobId(blobId)

	if loc, err = index.ReadOne(blobId); err != nil {
		err = errors.Wrapf(err, "Key: %s", formatKey)
		return
	}

	return
}

// TODO remove formatKey and switch to setting formatter on init
func (index *object_probe_index) ReadManyKeys(
	formatKey string,
	metadata *object_metadata.Metadata,
	locations *[]Loc,
) (err error) {
	var format object_inventory_format.Format

	if format, err = object_inventory_format.FormatForKeyError(
		formatKey,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var blobId interfaces.BlobId

	if blobId, err = object_inventory_format.GetDigestForMetadata(
		format,
		metadata,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.ReadMany(blobId, locations)
}

func (index *object_probe_index) ReadAll(
	metadata *object_metadata.Metadata,
	locations *[]Loc,
) (err error) {
	var blobIds map[string]interfaces.BlobId

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

func (index *object_probe_index) PrintAll(env env_ui.Env) (err error) {
	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]

		if err = page.PrintAll(env); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (index *object_probe_index) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()

	for pageIndex := range index.pages {
		page := &index.pages[pageIndex]
		wg.Do(page.Flush)
	}

	return wg.GetError()
}
