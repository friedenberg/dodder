package object_probe_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
)

type (
	Sha = sha.Sha

	commonInterface interface {
		// TODO rename to AddDigest and enforce digest type
		AddSha(interfaces.BlobId, Loc) error
		ReadOne(sh interfaces.BlobId) (loc Loc, err error)
		ReadMany(sh interfaces.BlobId, locs *[]Loc) (err error)
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
	pages [PageCount]page
}

func MakePermitDuplicates(
	envRepo env_repo.Env,
	path string,
) (index *object_probe_index, err error) {
	index = &object_probe_index{}
	err = index.initialize(rowEqualerComplete{}, envRepo, path)
	return
}

func MakeNoDuplicates(
	envRepo env_repo.Env,
	dir string,
) (index *object_probe_index, err error) {
	index = &object_probe_index{}
	err = index.initialize(rowEqualerShaOnly{}, envRepo, dir)
	return
}

func (index *object_probe_index) initialize(
	equaler interfaces.Equaler[*row],
	envRepo env_repo.Env,
	dir string,
) (err error) {
	for i := range index.pages {
		page := &index.pages[i]
		page.initialize(
			equaler,
			envRepo,
			page_id.PageIdFromPath(uint8(i), dir),
		)
	}

	return
}

func (index *object_probe_index) GetObjectProbeIndex() Index {
	return index
}

func (index *object_probe_index) AddMetadata(m *Metadata, loc Loc) (err error) {
	var shas map[string]interfaces.BlobId

	if shas, err = object_inventory_format.GetShasForMetadata(m); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, s := range shas {
		if err = index.addSha(s, loc); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (index *object_probe_index) AddSha(
	sh interfaces.BlobId,
	loc Loc,
) (err error) {
	return index.addSha(sh, loc)
}

func (index *object_probe_index) addSha(
	sh interfaces.BlobId,
	loc Loc,
) (err error) {
	if sh.IsNull() {
		return
	}

	var i uint8

	if i, err = page_id.PageIndexForDigest(DigitWidth, sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.pages[i].AddSha(sh, loc)
}

func (index *object_probe_index) ReadOne(
	sh interfaces.BlobId,
) (loc Loc, err error) {
	var i uint8

	if i, err = page_id.PageIndexForDigest(DigitWidth, sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.pages[i].ReadOne(sh)
}

func (index *object_probe_index) ReadMany(
	sh interfaces.BlobId,
	locs *[]Loc,
) (err error) {
	var i uint8

	if i, err = page_id.PageIndexForDigest(DigitWidth, sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.pages[i].ReadMany(sh, locs)
}

func (index *object_probe_index) ReadOneKey(
	kf string,
	m *object_metadata.Metadata,
) (loc Loc, err error) {
	var f object_inventory_format.FormatGeneric

	if f, err = object_inventory_format.FormatForKeyError(kf); err != nil {
		err = errors.Wrap(err)
		return
	}

	var sh interfaces.BlobId

	if sh, err = object_inventory_format.GetShaForMetadata(f, m); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer blob_ids.PutBlobId(sh)

	if loc, err = index.ReadOne(sh); err != nil {
		err = errors.Wrapf(err, "Key: %s", kf)
		return
	}

	return
}

func (index *object_probe_index) ReadManyKeys(
	kf string,
	m *object_metadata.Metadata,
	h *[]Loc,
) (err error) {
	var f object_inventory_format.FormatGeneric

	if f, err = object_inventory_format.FormatForKeyError(kf); err != nil {
		err = errors.Wrap(err)
		return
	}

	var sh interfaces.BlobId

	if sh, err = object_inventory_format.GetShaForMetadata(f, m); err != nil {
		err = errors.Wrap(err)
		return
	}

	return index.ReadMany(sh, h)
}

func (index *object_probe_index) ReadAll(
	m *object_metadata.Metadata,
	h *[]Loc,
) (err error) {
	var shas map[string]interfaces.BlobId

	if shas, err = object_inventory_format.GetShasForMetadata(m); err != nil {
		err = errors.Wrap(err)
		return
	}

	wg := errors.MakeWaitGroupParallel()

	for k, s := range shas {
		s := s
		wg.Do(
			func() (err error) {
				var loc Loc

				if loc, err = index.ReadOne(s); err != nil {
					err = errors.Wrapf(err, "Key: %s", k)
					return
				}

				*h = append(*h, loc)

				return
			},
		)
	}

	return wg.GetError()
}

func (index *object_probe_index) PrintAll(env env_ui.Env) (err error) {
	for i := range index.pages {
		p := &index.pages[i]

		if err = p.PrintAll(env); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (index *object_probe_index) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()

	for i := range index.pages {
		p := &index.pages[i]
		wg.Do(p.Flush)
	}

	return wg.GetError()
}
