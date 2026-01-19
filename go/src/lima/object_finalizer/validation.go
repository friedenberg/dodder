package object_finalizer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

func (finalizer *Finalizer) ValidateIfNecessary(
	daughter *sku.Transacted,
	mother *sku.Transacted,
	options sku.CommitOptions,
	// typedBlobStores typed_blob_store.Stores,
) (err error) {
	if !options.Validate {
		return err
	}

	switch daughter.GetSku().GetGenre() {
	case genres.Type:
		// var repool interfaces.FuncRepool

		// if _, repool, _, err = typedBlobStores.Type.ParseTypedBlob(
		// 	daughter.GetType(),
		// 	daughter.GetSku().GetBlobDigest(),
		// ); err != nil {
		// 	err = errors.Wrap(err)
		// 	return err
		// }

		// defer repool()
	}

	return err
}

type VerifyOptions struct {
	PubKeyPresent       bool
	ObjectDigestPresent bool
	ObjectSigPresent    bool
	ObjectSigValid      bool
}

func (finalizer *Finalizer) Verify(
	transacted *sku.Transacted,
) (err error) {
	pubKey := transacted.GetMetadata().GetRepoPubKey()

	if err = markl.AssertIdIsNotNull(
		pubKey,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.GetMetadata().GetObjectDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.GetMetadata().GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = pubKey.Verify(
		transacted.GetMetadata().GetObjectDigest(),
		transacted.GetMetadata().GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
