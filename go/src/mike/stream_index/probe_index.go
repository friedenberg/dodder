package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_map"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
	"code.linenisgreat.com/dodder/go/src/lima/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

type probeIndex struct {
	defaultObjectDigestMarklFormatId string
	index                            *object_probe_index.Index
	additionProbes                   collections_map.Map[string, *sku.Transacted]
}

func (index *Index) PrintAllProbes() (err error) {
	if index.probeIndex.index.PrintAll(index.envRepo); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (index *probeIndex) Initialize(
	envRepo env_repo.Env,
	hashType markl.FormatHash,
) (err error) {
	index.defaultObjectDigestMarklFormatId = envRepo.GetObjectDigestType()

	if index.index, err = object_probe_index.MakeNoDuplicates(
		envRepo,
		envRepo.DirIndexObjectPointers(),
		hashType,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	index.additionProbes = make(collections_map.Map[string, *sku.Transacted])

	return err
}

func (index *probeIndex) Flush() (err error) {
	if err = index.index.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	index.additionProbes.Reset()

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
	for probeId := range object.AllProbeIds(
		index.index.GetHashType(),
		index.defaultObjectDigestMarklFormatId,
	) {
		if err = index.index.AddDigest(
			ids.ProbeIdWithObjectId{
				ObjectId: object.GetObjectId(),
				ProbeId:  probeId,
			},
			loc,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
