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
	return transacted.calculateObjectSha(false, false)
}

type funcCalcDigest func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.BlobId, error)

func (transacted *Transacted) calculateObjectDigest(
	funcCalcDigest funcCalcDigest,
) (err error) {
	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.makeDigestCalcFunc(
		funcCalcDigest,
		object_inventory_format.FormatV11ObjectDigest,
		transacted.Metadata.GetObjectDigestMutable(),
	)(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) makeDigestCalcFunc(
	funcCalcDigest funcCalcDigest,
	objectFormat object_inventory_format.Format,
	digest interfaces.MutableBlobId,
) errors.FuncErr {
	return func() (err error) {
		var actual interfaces.BlobId

		if actual, err = funcCalcDigest(
			objectFormat,
			transacted,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer merkle_ids.PutBlobId(actual)

		if err = digest.SetMerkleId(
			merkle.HRPObjectDigestSha256V1,
			actual.GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (transacted *Transacted) makeShaCalcFunc(
	funcCalcDigest funcCalcDigest,
	objectFormat object_inventory_format.Format,
	sh *sha.Sha,
) errors.FuncErr {
	return func() (err error) {
		var actual interfaces.BlobId

		if actual, err = funcCalcDigest(
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

func (transacted *Transacted) calculateObjectSha(
	debug bool,
	includeMerkle bool,
) (err error) {
	funcCalcDigest := object_inventory_format.GetDigestForContext

	if debug {
		funcCalcDigest = object_inventory_format.GetDigestForContextDebug
	}

	wg := errors.MakeWaitGroupParallel()

	wg.Do(
		transacted.makeShaCalcFunc(
			funcCalcDigest,
			object_inventory_format.FormatsV5MetadataSansTai,
			&transacted.Metadata.SelfWithoutTai,
		),
	)

	if includeMerkle {
		wg.Do(func() error {
			return transacted.calculateMerkleObjectDigest(funcCalcDigest)
		})
	}

	wg.Do(func() error {
		return transacted.calculateObjectDigest(funcCalcDigest)
	})

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) calculateMerkleObjectDigest(
	funcCalcDigest funcCalcDigest,
) (err error) {
	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO enforce that existing object digests are explicitly discarded before
	// overwriting
	// if err = merkle_ids.MakeErrIsNotNull(
	// 	transacted.Metadata.GetObjectDigest(),
	// ); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	if err = transacted.makeDigestCalcFunc(
		funcCalcDigest,
		object_inventory_format.FormatV11ObjectDigest,
		transacted.Metadata.GetObjectDigestMutable(),
	)(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

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
	pubKey := transacted.Metadata.GetRepoPubKey()

	if err = merkle_ids.MakeErrIsNull(
		pubKey,
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle_ids.MakeErrIsNull(
		transacted.Metadata.GetObjectSig(),
		"object-sig",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.CalculateObjectDigests(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle.VerifySignature(
		pubKey.GetBytes(),
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
