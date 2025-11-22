package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/golf/page_id"
	"code.linenisgreat.com/dodder/go/src/lima/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func (index *Index) ReadOneMarklIdAdded(
	marklId interfaces.MarklId,
	object *sku.Transacted,
) (ok bool) {
	additionObject, ok := index.additionProbes.Get(string(marklId.GetBytes()))

	if ok {
		sku.TransactedResetter.ResetWith(object, additionObject)
		return ok
	}

	return ok
}

// TODO migrate to panic semantics
func (index *Index) ReadOneMarklId(
	marklId interfaces.MarklId,
	object *sku.Transacted,
) (err error) {
	errors.PanicIfError(markl.AssertIdIsNotNull(marklId))

	var loc object_probe_index.Loc

	if loc, err = index.readOneMarklIdLoc(marklId); err != nil {
		return err
	}

	// TODO read from page additions if necessary
	if err = index.readOneLoc(loc, object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (index *Index) ReadManyMarklId(
	marklId interfaces.MarklId,
) (objects []*sku.Transacted, err error) {
	// TODO read from page additions if necessary
	var locs []object_probe_index.Loc

	if locs, err = index.readManyMarklIdLoc(marklId); err != nil {
		err = errors.Wrap(err)
		return objects, err
	}

	for _, loc := range locs {
		object := sku.GetTransactedPool().Get()

		if err = index.readOneLoc(loc, object); err != nil {
			err = errors.Wrapf(err, "Loc: %s", loc)
			return objects, err
		}

		objects = append(objects, object)
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

	if page.objectIdStringExists(objectIdString) {
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
	pageReader, pageReaderClose := index.makeProbePageReader(loc.Page)
	defer errors.Deferred(&err, pageReaderClose)

	if err = pageReader.readOneCursor(loc.Cursor, object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
