package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/golf/page_id"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/lima/object_probe_index"
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

func (index *Index) ReadOneMarklId(
	marklId interfaces.MarklId,
	object *sku.Transacted,
) (ok bool) {
	errors.PanicIfError(markl.AssertIdIsNotNull(marklId))

	var loc object_probe_index.Loc

	{
		var err error

		// TODO migrate to panic semantics
		if loc, err = index.readOneMarklIdLoc(marklId); err != nil {
			if errors.IsNotExist(err) || collections.IsErrNotFound(err) {
				return ok
			} else {
				panic(err)
			}
		}
	}

	ok = index.readOneLoc(loc, object)

	return ok
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

		if !index.readOneLoc(loc, object) {
			err = errors.Errorf("failed to read loc: %s", loc)
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
	defer markl.PutId(digest)

	if _, err = index.readOneMarklIdLoc(digest); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (index *Index) ReadOneObjectId(
	objectId ids.ObjectIdLike,
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

	if !index.ReadOneMarklId(digest, object) {
		err = collections.MakeErrNotFoundString(objectIdString)
		return err
	}

	return err
}

func (index *Index) ReadManyObjectId(
	objectId interfaces.ObjectId,
) (objects []*sku.Transacted, err error) {
	digest := markl.FormatHashSha256.FromStringContent(objectId.String())
	defer markl.PutId(digest)

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

	key := objectId.String() + tai.String()

	digest := markl.FormatHashSha256.FromStringContent(key)
	defer markl.PutId(digest)

	object = sku.GetTransactedPool().Get()

	if !index.ReadOneMarklId(digest, object) {
		err = collections.MakeErrNotFoundString(key)
		return object, err
	}

	return object, err
}

func (index *Index) readOneLoc(
	loc object_probe_index.Loc,
	object *sku.Transacted,
) (ok bool) {
	pageReader, pageReaderClose := index.makeProbePageReader(loc.Page)
	defer errors.Must(pageReaderClose)

	ok = pageReader.readOneCursor(loc.Cursor, object)

	return ok
}
