package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
)

type funcCalcDigest func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.MarklId, error)

// calculates the respective digests
func (transacted *Transacted) finalize(
	debug bool,
	includeMerkle bool,
) (err error) {
	funcCalcDigest := object_inventory_format.GetDigestForContext

	if debug {
		funcCalcDigest = object_inventory_format.GetDigestForContextDebug
	}

	wg := errors.MakeWaitGroupParallel()

	wg.Do(
		transacted.makeDigestCalcFunc(
			funcCalcDigest,
			object_inventory_format.FormatsV5MetadataSansTai,
			&transacted.Metadata.SelfWithoutTai,
			markl.HRPObjectDigestSha256V1,
		),
	)

	if includeMerkle {
		wg.Do(func() error {
			return transacted.calculateObjectDigestMerkle(funcCalcDigest)
		})
	}

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) makeDigestCalcFunc(
	funcCalcDigest funcCalcDigest,
	objectFormat object_inventory_format.Format,
	digest interfaces.MutableMarklId,
	tipe string,
) errors.FuncErr {
	return func() (err error) {
		var actual interfaces.MarklId

		if actual, err = funcCalcDigest(
			objectFormat,
			transacted,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer markl.PutBlobId(actual)

		if err = digest.SetMerkleId(
			tipe,
			actual.GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (transacted *Transacted) CalculateObjectDigestSelfWithTai(
	funcCalcDigest funcCalcDigest,
) (err error) {
	if err = transacted.makeDigestCalcFunc(
		funcCalcDigest,
		object_inventory_format.FormatsV5MetadataSansTai,
		&transacted.Metadata.SelfWithoutTai,
		markl.HRPObjectDigestSha256V1,
	)(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) calculateObjectDigestMerkle(
	funcCalcDigest funcCalcDigest,
) (err error) {
	if err = markl.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.MakeErrIsNotNull(
		transacted.Metadata.GetObjectDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.makeDigestCalcFunc(
		funcCalcDigest,
		object_inventory_format.FormatV11ObjectDigest,
		transacted.Metadata.GetObjectDigestMutable(),
		markl.HRPObjectDigestSha256V1,
	)(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
