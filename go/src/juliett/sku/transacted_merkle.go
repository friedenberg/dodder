package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
)

// TODO replace this with repo signatures
// TODO include repo pubkey
func (transacted *Transacted) CalculateObjectDigests() (err error) {
	return transacted.calculateObjectSha(false)
}

func (transacted *Transacted) calculateObjectDigest() (err error) {
	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.makeDigestCalcFunc(
		object_inventory_format.GetDigestForContext,
		object_inventory_format.FormatV11ObjectDigest,
		transacted.Metadata.GetObjectDigestMutable(),
	)(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) makeDigestCalcFunc(
	funkMakeBlobId func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.BlobId, error),
	objectFormat object_inventory_format.Format,
	digest interfaces.MutableBlobId,
) errors.FuncErr {
	return func() (err error) {
		var actual interfaces.BlobId

		if actual, err = funkMakeBlobId(
			objectFormat,
			transacted,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer merkle_ids.PutBlobId(actual)

		if err = digest.SetMerkleId(merkle.HRPObjectDigestSha256V1, actual.GetBytes()); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (transacted *Transacted) makeShaCalcFunc(
	f func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.BlobId, error),
	objectFormat object_inventory_format.Format,
	sh *sha.Sha,
) errors.FuncErr {
	return func() (err error) {
		var actual interfaces.BlobId

		if actual, err = f(
			objectFormat,
			transacted,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer merkle_ids.PutBlobId(actual)

		if err = sh.SetDigest(actual); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (transacted *Transacted) calculateObjectSha(debug bool) (err error) {
	f := object_inventory_format.GetDigestForContext

	if debug {
		f = object_inventory_format.GetDigestForContextDebug
	}

	wg := errors.MakeWaitGroupParallel()
	wg.Do(transacted.calculateObjectDigest)

	wg.Do(
		transacted.makeShaCalcFunc(
			f,
			object_inventory_format.FormatsV5MetadataSansTai,
			&transacted.Metadata.SelfWithoutTai,
		),
	)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO turn into proper merkle tree
func (transacted *Transacted) SignOverwrite(
	config genesis_configs.ConfigPrivate,
) (err error) {
	transacted.CalculateObjectDigests()

	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	privateKey := config.GetPrivateKey()

	var bites []byte

	if bites, err = merkle.Sign(
		privateKey,
		transacted.Metadata.GetObjectDigest().GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.Metadata.GetObjectSigMutable().SetMerkleId(
		merkle.HRPRepoSigV1,
		bites,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) Verify() (err error) {
	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrapf(
			err,
			"Object: %s, Fields: %#v",
			String(transacted),
			transacted.Metadata.Fields,
		)
		return
	}

	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetObjectSig(),
		"object-sig",
	); err != nil {
		err = errors.Wrapf(
			err,
			"Object: %s, Fields: %#v",
			String(transacted),
			transacted.Metadata.Fields,
		)
		return
	}

	if err = transacted.CalculateObjectDigests(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle.VerifySignature(
		transacted.Metadata.GetRepoPubKey().GetBytes(),
		transacted.Metadata.GetObjectDigestMutable().GetBytes(),
		transacted.Metadata.GetObjectSig().GetBytes(),
	); err != nil {
		err = errors.Wrapf(
			err,
			"Sku: %s, Tags %s",
			String(transacted),
			quiter.StringCommaSeparated(transacted.Metadata.GetTags()),
		)
		return
	}

	return
}
