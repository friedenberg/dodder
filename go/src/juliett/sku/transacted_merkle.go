package sku

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

func (transacted *Transacted) SetMother(mother *Transacted) (err error) {
	motherSig := transacted.Metadata.GetMotherObjectSigMutable()

	if mother == nil {
		motherSig.Reset()
		return
	}

	if err = motherSig.SetMerkleId(
		merkle.HRPObjectMotherSigV1,
		mother.Metadata.GetObjectSig().GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// calculates the object digests using the object's repo pubkey
func (transacted *Transacted) FinalizeUsingObject() (err error) {
	if err = merkle.MakeErrIsNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return transacted.FinalizeUsingRepoPubKey(
		transacted.Metadata.GetRepoPubKey().GetBytes(),
	)
}

// calculates the object digests using the provided repo pubkey
func (transacted *Transacted) FinalizeUsingRepoPubKey(
	pubKey merkle.PublicKey,
) (err error) {
	pubKeyMutable := transacted.Metadata.GetRepoPubKeyMutable()

	if pubKeyMutable.IsNull() {
		if err = pubKeyMutable.SetMerkleId(
			merkle.HRPRepoPubKeyV1,
			pubKey,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = merkle.MakeErrNotEqualBytes(
			pubKey,
			pubKeyMutable.GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = transacted.finalize(false, true); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO remove / rename
func (transacted *Transacted) CalculateObjectDigests() (err error) {
	return transacted.finalize(false, false)
}

func (transacted *Transacted) FinalizeAndSignIfNecessary(
	config genesis_configs.ConfigPrivate,
) (err error) {
	if !transacted.Metadata.GetObjectSig().IsNull() {
		return
	}

	if err = transacted.FinalizeAndSign(config); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) FinalizeAndSignOverwrite(
	config genesis_configs.ConfigPrivate,
) (err error) {
	transacted.Metadata.GetObjectSigMutable().Reset()
	transacted.Metadata.GetRepoPubKeyMutable().Reset()

	if err = transacted.FinalizeAndSign(
		config,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) FinalizeAndSign(
	config genesis_configs.ConfigPrivate,
) (err error) {
	if err = merkle.MakeErrIsNotNull(
		transacted.Metadata.GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle.MakeErrIsNotNull(
		transacted.Metadata.GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.Metadata.GetRepoPubKeyMutable().SetMerkleId(
		merkle.HRPRepoPubKeyV1,
		config.GetPublicKey(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.FinalizeUsingObject(); err != nil {
		err = errors.Wrap(err)
		return
	}

	privateKey := config.GetPrivateKey()

	var bites []byte

	if bites, err = merkle.SignBytes(
		privateKey,
		transacted.Metadata.GetObjectDigest().GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.Metadata.GetObjectSigMutable().SetMerkleId(
		config.GetObjectSigTypeString(),
		bites,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = transacted.Verify(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) FinalizeAndVerify() (err error) {
	if err = transacted.FinalizeUsingObject(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if slices.Contains(
		[]string{ids.TypeInventoryListV1},
		transacted.GetType().String(),
	) {
		return
	}

	if err = transacted.Verify(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) Verify() (err error) {
	pubKey := transacted.Metadata.GetRepoPubKey()

	if err = merkle.MakeErrIsNull(
		pubKey,
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle.MakeErrIsNull(
		transacted.Metadata.GetObjectDigest(),
		"object-digest",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle.MakeErrIsNull(
		transacted.Metadata.GetObjectSig(),
		"object-sig",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = merkle.VerifyBytes(
		pubKey.GetBytes(),
		transacted.Metadata.GetObjectDigest().GetBytes(),
		transacted.Metadata.GetObjectSig().GetBytes(),
	); err != nil {
		err = errors.Wrapf(err, "Object: %q", String(transacted))
		return
	}

	return
}
