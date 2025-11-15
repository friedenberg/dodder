package object_finalizer

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	FinalizerGetter interface {
		GetObjectFinalizer() Finalizer
	}

	Finalizer = finalizer

	finalizer struct {
		pubKey interfaces.MarklId
	}
)

func (finalizer finalizer) GetObjectFinalizer() Finalizer {
	return finalizer
}

// TODO extract into a versioned object finalizer
// calculates the object digests using the object's repo pubkey
func (finalizer finalizer) FinalizeUsingObject(transacted *sku.Transacted) (err error) {
	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = finalizer.FinalizeUsingRepoPubKey(
		transacted,
		transacted.Metadata.GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// calculates the object digests using the provided repo pubkey
func (finalizer finalizer) FinalizeUsingRepoPubKey(
	transacted *sku.Transacted,
	pubKey interfaces.MarklId,
) (err error) {
	// TODO migrate this to config
	pubKeyMutable := transacted.Metadata.GetRepoPubKeyMutable()

	if pubKeyMutable.IsNull() {
		pubKeyMutable.ResetWithMarklId(pubKey)
		// if err = markl.SetMerkleIdWithFormat(
		// 	transacted.Metadata.GetRepoPubKeyMutable(),
		// 	markl.FormatIdRepoPubKeyV1,
		// 	pubKey,
		// ); err != nil {
		// 	err = errors.Wrap(err)
		// 	return
		// }
	} else {
		if err = markl.MakeErrNotEqualBytes(
			pubKey.GetBytes(),
			pubKeyMutable.GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	// TODO populate lockfile
	// read current signature of type
	// write to lock

	if err = transacted.CalculateDigests(
		false,
		transacted.GetDigestWriteMapWithMerkle(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (finalizer finalizer) FinalizeWithoutPubKey(
	transacted *sku.Transacted,
) (err error) {
	transacted.Metadata.GetRepoPubKeyMutable().Reset()

	if err = transacted.CalculateDigests(
		false,
		transacted.GetDigestWriteMapWithMerkle(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO remove / rename
func (finalizer finalizer) CalculateObjectDigests(
	transacted *sku.Transacted,
) (err error) {
	return transacted.CalculateDigests(
		false,
		transacted.GetDigestWriteMapWithoutMerkle(),
	)
}

func (finalizer finalizer) FinalizeAndSignIfNecessary(
	transacted *sku.Transacted,
	config genesis_configs.ConfigPrivate,
) (err error) {
	if !transacted.Metadata.GetObjectSig().IsNull() {
		return err
	}

	if err = finalizer.FinalizeAndSign(transacted, config); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if transacted.Metadata.GetRepoPubKey().GetPurpose() == "" {
		panic("empty pbukey format")
	}

	return err
}

func (finalizer finalizer) FinalizeAndSignOverwrite(
	transacted *sku.Transacted,
	config genesis_configs.ConfigPrivate,
) (err error) {
	// TODO populate format ids from config
	transacted.Metadata.GetObjectSigMutable().Reset()
	transacted.Metadata.GetRepoPubKeyMutable().Reset()

	if err = finalizer.FinalizeAndSign(
		transacted,
		config,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (finalizer finalizer) FinalizeAndSign(
	transacted *sku.Transacted,
	config genesis_configs.ConfigPrivate,
) (err error) {
	if err = markl.AssertIdIsNull(
		transacted.Metadata.GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNull(
		transacted.Metadata.GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	transacted.Metadata.GetRepoPubKeyMutable().ResetWithMarklId(
		config.GetPublicKey(),
	)

	if err = finalizer.FinalizeUsingObject(transacted); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectDigest()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	privateKey := config.GetPrivateKey()

	if err = privateKey.Sign(
		transacted.Metadata.GetObjectDigest(),
		transacted.Metadata.GetObjectSigMutable(),
		config.GetObjectSigMarklTypeId(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = transacted.Verify(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (finalizer finalizer) FinalizeAndVerify(
	transacted *sku.Transacted,
) (err error) {
	if err = finalizer.FinalizeUsingObject(transacted); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if slices.Contains(
		[]string{ids.TypeInventoryListV1},
		transacted.GetType().String(),
	) {
		return err
	}

	if err = transacted.Verify(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
