package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
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
	for probeId := range object.AllProbeIds(index.hashType) {
		// ui.Debug().Print(
		// 	probeId.Key,
		// 	probeId.Id.StringWithFormat(),
		// 	sku.String(object),
		// )
		if err = index.index.AddDigest(probeId.Id, loc); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (index *Index) ReadOneMarklId(
	blobId interfaces.MarklId,
	object *sku.Transacted,
) (err error) {
	errors.PanicIfError(markl.AssertIdIsNotNull(blobId))

	originalBlobId := blobId

	if blobId.GetMarklFormat().GetMarklFormatId() != index.hashType.GetMarklFormatId() {
		replacementId, repool := index.hashType.GetMarklIdForMarklId(blobId)
		defer repool()

		blobId = replacementId
	}

	var loc object_probe_index.Loc

	if loc, err = index.readOneMarklIdLoc(blobId); err != nil {
		if originalBlobId != blobId {
			// ui.Debug().Print(
			// 	originalBlobId.StringWithFormat(),
			// 	blobId.StringWithFormat(),
			// )
		}
		return err
	}

	if err = index.readOneLoc(loc, object); err != nil {
		return err
	}

	return err
}

func (index *Index) ReadManyMarklId(
	blobId interfaces.MarklId,
) (objects []*sku.Transacted, err error) {
	var locs []object_probe_index.Loc

	if locs, err = index.readManyMarklIdLoc(blobId); err != nil {
		err = errors.Wrap(err)
		return objects, err
	}

	for _, loc := range locs {
		sk := sku.GetTransactedPool().Get()

		if err = index.readOneLoc(loc, sk); err != nil {
			err = errors.Wrapf(err, "Loc: %s", loc)
			return objects, err
		}

		objects = append(objects, sk)
	}

	return objects, err
}

func (index *Index) ObjectExists(
	objectId *ids.ObjectId,
) (err error) {
	var pageIndex uint8

	objectIdString := objectId.String()

	if pageIndex, err = page_id.PageIndexForString(
		DigitWidth,
		objectIdString,
		index.hashType,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	page := index.GetPage(pageIndex)

	if _, ok := page.addedObjectIdLookup[objectIdString]; ok {
		return err
	}

	digest := index.hashType.FromStringContent(objectIdString)
	defer markl.PutBlobId(digest)

	if _, err = index.readOneMarklIdLoc(digest); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (index *Index) ReadOneObjectId(
	objectId interfaces.ObjectId,
	object *sku.Transacted,
) (err error) {
	objectIdString := objectId.String()

	if objectIdString == "" {
		panic("empty object id")
	}

	digest, repool := markl.FormatHashSha256.GetMarklIdForString(
		objectIdString,
	)
	defer repool()

	if err = index.ReadOneMarklId(digest, object); err != nil {
		return err
	}

	return err
}

func (index *Index) ReadManyObjectId(
	objectId interfaces.ObjectId,
) (objects []*sku.Transacted, err error) {
	digest := markl.FormatHashSha256.FromStringContent(objectId.String())
	defer markl.PutBlobId(digest)

	if objects, err = index.ReadManyMarklId(digest); err != nil {
		err = errors.Wrap(err)
		return objects, err
	}

	return objects, err
}

// TODO switch to empty=not found semantics instead of error
func (index *Index) ReadOneObjectIdTai(
	objectId interfaces.ObjectId,
	tai ids.Tai,
) (object *sku.Transacted, err error) {
	if tai.IsEmpty() {
		err = collections.MakeErrNotFoundString(tai.String())
		return object, err
	}

	digest := markl.FormatHashSha256.FromStringContent(
		objectId.String() + tai.String(),
	)
	defer markl.PutBlobId(digest)

	object = sku.GetTransactedPool().Get()

	if err = index.ReadOneMarklId(digest, object); err != nil {
		return object, err
	}

	return object, err
}

func (index *Index) readOneLoc(
	loc object_probe_index.Loc,
	object *sku.Transacted,
) (err error) {
	page := index.pages[loc.Page]

	if err = page.readOneRange(loc.Cursor, object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
