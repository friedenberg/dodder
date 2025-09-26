package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) ReadTransactedFromObjectId(
	k1 interfaces.ObjectId,
) (sk1 *sku.Transacted, err error) {
	sk1 = sku.GetTransactedPool().Get()

	if err = store.ReadOneInto(k1, sk1); err != nil {
		if collections.IsErrNotFound(err) {
			sku.GetTransactedPool().Put(sk1)
			sk1 = nil
		}

		err = errors.Wrap(err)
		return sk1, err
	}

	return sk1, err
}

// TODO transition to a context-based panic / cancel semantic
func (store *Store) ReadOneObjectId(
	objectId interfaces.ObjectId,
) (object *sku.Transacted, err error) {
	if objectId.IsEmpty() {
		return object, err
	}

	object = sku.GetTransactedPool().Get()

	if err = store.streamIndex.ReadOneObjectId(objectId, object); err != nil {
		if !collections.IsErrNotFound(err) {
			err = errors.Wrap(err)
		}

		return object, err
	}

	return object, err
}

// TODO add support for cwd and sigil
// TODO simplify
func (store *Store) ReadOneInto(
	objectId interfaces.ObjectId,
	out *sku.Transacted,
) (err error) {
	var sk *sku.Transacted

	switch objectId.GetGenre() {
	case genres.Zettel:
		var zettelId *ids.ZettelId

		if zettelId, err = store.GetAbbrStore().GetZettelIds().ExpandString(
			objectId.String(),
		); err == nil {
			objectId = zettelId
		} else {
			err = nil
		}

		if sk, err = store.ReadOneObjectId(objectId); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case genres.Type, genres.Tag, genres.Repo, genres.InventoryList:
		if sk, err = store.ReadOneObjectId(objectId); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case genres.Config:
		sk = store.GetConfigStore().GetConfig().GetSku()

		if sk.GetTai().IsEmpty() {
			ui.Err().Print("config tai is empty")
		}

	case genres.Blob:
		var oid ids.ObjectId

		if err = oid.SetWithIdLike(objectId); err != nil {
			err = collections.MakeErrNotFound(objectId)
			return err
		}

		if sk, err = store.ReadOneObjectId(objectId); err != nil {
			err = errors.Wrap(err)
			return err
		}

	default:
		err = genres.MakeErrUnsupportedGenre(objectId)
		return err
	}

	if sk == nil {
		err = collections.MakeErrNotFound(objectId)
		return err
	}

	sku.TransactedResetter.ResetWith(out, sk)

	return err
}

func (store *Store) ReadPrimitiveQuery(
	query sku.PrimitiveQueryGroup,
	funcIter interfaces.FuncIter[*sku.Transacted],
) (err error) {
	return store.streamIndex.ReadPrimitiveQuery(query, funcIter)
}
