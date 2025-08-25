package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type probe_index struct {
	envRepo env_repo.Env
	object_probe_index.Index
}

func (index *probe_index) Initialize(
	envRepo env_repo.Env,
) (err error) {
	index.envRepo = envRepo

	if index.Index, err = object_probe_index.MakeNoDuplicates(
		index.envRepo,
		index.envRepo.DirCacheObjectPointers(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (index *probe_index) Flush() (err error) {
	if err = index.Index.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (index *probe_index) readOneShaLoc(
	sh interfaces.BlobId,
) (loc object_probe_index.Loc, err error) {
	if loc, err = index.Index.ReadOne(sh); err != nil {
		return
	}

	return
}

func (index *probe_index) readManyShaLoc(
	sh interfaces.BlobId,
) (locs []object_probe_index.Loc, err error) {
	if err = index.Index.ReadMany(sh, &locs); err != nil {
		return
	}

	return
}

func (index *probe_index) saveOneLoc(
	o *sku.Transacted,
	loc object_probe_index.Loc,
) (err error) {
	if err = index.saveOneLocString(
		o.GetObjectId().String(),
		loc,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = index.saveOneLocString(
		o.GetObjectId().String()+o.GetTai().String(),
		loc,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (index *probe_index) saveOneLocString(
	str string,
	loc object_probe_index.Loc,
) (err error) {
	digest := sha.FromStringContent(str)
	defer merkle_ids.PutBlobId(digest)

	if err = index.Index.AddSha(digest, loc); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
