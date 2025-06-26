package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type probe_index struct {
	directoryLayout env_repo.Env
	object_probe_index.Index
}

func (s *probe_index) Initialize(
	directoryLayout env_repo.Env,
) (err error) {
	s.directoryLayout = directoryLayout

	if s.Index, err = object_probe_index.MakeNoDuplicates(
		s.directoryLayout,
		s.directoryLayout.DirCacheObjectPointers(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *probe_index) Flush() (err error) {
	if err = s.Index.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *probe_index) readOneShaLoc(
	sh *sha.Sha,
) (loc object_probe_index.Loc, err error) {
	if loc, err = s.Index.ReadOne(sh); err != nil {
		return
	}

	return
}

func (s *probe_index) readManyShaLoc(
	sh *sha.Sha,
) (locs []object_probe_index.Loc, err error) {
	if err = s.Index.ReadMany(sh, &locs); err != nil {
		return
	}

	return
}

func (s *probe_index) saveOneLoc(
	o *sku.Transacted,
	loc object_probe_index.Loc,
) (err error) {
	if err = s.saveOneLocString(
		o.GetObjectId().String(),
		loc,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.saveOneLocString(
		o.GetObjectId().String()+o.GetTai().String(),
		loc,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *probe_index) saveOneLocString(
	str string,
	loc object_probe_index.Loc,
) (err error) {
	sh := sha.FromStringContent(str)
	defer sha.GetPool().Put(sh)

	if err = s.Index.AddSha(sh, loc); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
