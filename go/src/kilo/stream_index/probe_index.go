package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type probeIndex struct {
	envRepo  env_repo.Env
	index    *object_probe_index.Index
	hashType markl.FormatHash
}

func (index *probeIndex) Initialize(
	envRepo env_repo.Env,
	hashType markl.FormatHash,
) (err error) {
	index.envRepo = envRepo
	index.hashType = hashType

	if index.index, err = object_probe_index.MakeNoDuplicates(
		index.envRepo,
		index.envRepo.DirCacheObjectPointers(),
		index.hashType,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (index *probeIndex) Flush() (err error) {
	if err = index.index.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (index *probeIndex) readOneMarklIdLoc(
	blobId interfaces.MarklId,
) (loc object_probe_index.Loc, err error) {
	if loc, err = index.index.ReadOne(blobId); err != nil {
		return loc, err
	}

	return loc, err
}

func (index *probeIndex) readManyMarklIdLoc(
	blobId interfaces.MarklId,
) (locs []object_probe_index.Loc, err error) {
	if err = index.index.ReadMany(blobId, &locs); err != nil {
		return locs, err
	}

	return locs, err
}

func (index *probeIndex) saveOneObjectLoc(
	object *sku.Transacted,
	loc object_probe_index.Loc,
) (err error) {
	for probeId := range object.AllProbeIds() {
		if err = index.index.AddDigest(probeId.Id, loc); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
