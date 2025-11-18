package store_browser

import (
	"net/url"

	"code.linenisgreat.com/dodder/go/src/_/external_state"
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/papa/queries"
)

type executor struct {
	store *Store
	qg    *queries.Query
	out   interfaces.FuncIter[sku.SkuType]
	co    sku.CheckedOut
}

func (c *executor) tryToEmitOneExplicitlyCheckedOut(
	internal *sku.Transacted,
	item Item,
) (err error) {
	c.co.GetSkuExternal().ObjectId.Reset()

	var uSku *url.URL

	if uSku, err = c.store.getUrl(internal); err != nil {
		err = errors.Wrap(err)
		return err
	}

	sku.TransactedResetter.ResetWith(c.co.GetSku(), internal)
	sku.TransactedResetter.ResetWith(c.co.GetSkuExternal().GetSku(), internal)

	if *uSku == item.Url.Url() {
		// c.co.SetState(checked_out_state.ExistsAndSame)
	} else {
		// c.co.SetState(checked_out_state.Changed)
	}

	c.co.GetSkuExternal().State = external_state.Tracked

	if err = c.tryToEmitOneCommon(item); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (c *executor) tryToEmitOneRecognized(
	internal *sku.Transacted,
	item Item,
) (err error) {
	c.co.SetState(checked_out_state.Recognized)

	if !queries.ContainsSkuCheckedOutState(c.qg, c.co.GetState()) {
		return err
	}

	sku.TransactedResetter.ResetWith(c.co.GetSku(), internal)
	sku.TransactedResetter.ResetWith(c.co.GetSkuExternal().GetSku(), internal)

	// if err = item.WriteToObjectId(&c.co.External.ObjectId); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	c.co.SetState(checked_out_state.Recognized)
	c.co.GetSkuExternal().State = external_state.Recognized

	if err = c.tryToEmitOneCommon(item); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (c *executor) tryToEmitOneUntracked(
	item Item,
) (err error) {
	c.co.SetState(checked_out_state.Untracked)

	if !queries.ContainsSkuCheckedOutState(c.qg, c.co.GetState()) {
		return err
	}

	sku.TransactedResetter.Reset(c.co.GetSkuExternal().GetSku())
	sku.TransactedResetter.Reset(c.co.GetSku())

	if err = c.co.GetSkuExternal().Metadata.Description.Set(item.Title); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = c.tryToEmitOneCommon(item); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (c *executor) tryToEmitOneCommon(
	i Item,
) (err error) {
	external := c.co.GetSkuExternal()

	if err = i.WriteToExternal(external); err != nil {
		err = errors.Wrap(err)
		return err
	}

	external.ObjectId.SetGenre(genres.Zettel)
	external.ExternalObjectId.SetGenre(genres.Zettel)

	if !queries.ContainsExternalSku(c.qg, external, c.co.GetState()) {
		return err
	}

	c.co.GetSkuExternal().RepoId = c.store.externalStoreInfo.RepoId

	if err = c.out(&c.co); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
