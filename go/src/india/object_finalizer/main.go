package object_finalizer

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

type (
	FinalizerGetter interface {
		GetObjectFinalizer() Finalizer
	}

	Finalizer = finalizer

	finalizer struct {
		// pubKey interfaces.MarklId
	}

	ObjectDigestWriteMap map[string]interfaces.MutableMarklId

	object interface {
		object_metadata.GetterMutable

		CalculateDigests(
			debug bool,
			formats ObjectDigestWriteMap,
		) (err error)

		GetDigestWriteMapWithMerkle() ObjectDigestWriteMap
		GetDigestWriteMapWithoutMerkle() ObjectDigestWriteMap
		Verify() (err error)
	}
)

func Make() Finalizer {
	return finalizer{}
}

func (finalizer finalizer) GetObjectFinalizer() Finalizer {
	return finalizer
}

// TODO extract into a versioned object finalizer
// calculates the object digests using the object's repo pubkey
func (finalizer finalizer) FinalizeUsingObject(transacted object) (err error) {
	if err = markl.AssertIdIsNotNull(
		transacted.GetMetadataMutable().GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = finalizer.FinalizeUsingRepoPubKey(
		transacted,
		transacted.GetMetadataMutable().GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// calculates the object digests using the provided repo pubkey
func (finalizer finalizer) FinalizeUsingRepoPubKey(
	transacted object,
	pubKey interfaces.MarklId,
) (err error) {
	// TODO migrate this to config
	pubKeyMutable := transacted.GetMetadataMutable().GetRepoPubKeyMutable()

	if pubKeyMutable.IsNull() {
		pubKeyMutable.ResetWithMarklId(pubKey)
		// if err = markl.SetMerkleIdWithFormat(
		// 	transacted.GetMetadataMutable().GetRepoPubKeyMutable(),
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
	transacted object,
) (err error) {
	transacted.GetMetadataMutable().GetRepoPubKeyMutable().Reset()

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
	transacted object,
) (err error) {
	return transacted.CalculateDigests(
		false,
		transacted.GetDigestWriteMapWithoutMerkle(),
	)
}

func (finalizer finalizer) FinalizeAndSignIfNecessary(
	transacted object,
	config genesis_configs.ConfigPrivate,
) (err error) {
	if !transacted.GetMetadataMutable().GetObjectSig().IsNull() {
		return err
	}

	if err = finalizer.FinalizeAndSign(transacted, config); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if transacted.GetMetadataMutable().GetRepoPubKey().GetPurpose() == "" {
		panic("empty pbukey format")
	}

	return err
}

func (finalizer finalizer) FinalizeAndSignOverwrite(
	transacted object,
	config genesis_configs.ConfigPrivate,
) (err error) {
	// TODO populate format ids from config
	transacted.GetMetadataMutable().GetObjectSigMutable().Reset()
	transacted.GetMetadataMutable().GetRepoPubKeyMutable().Reset()

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
	transacted object,
	config genesis_configs.ConfigPrivate,
) (err error) {
	if err = markl.AssertIdIsNull(
		transacted.GetMetadataMutable().GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNull(
		transacted.GetMetadataMutable().GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	transacted.GetMetadataMutable().GetRepoPubKeyMutable().ResetWithMarklId(
		config.GetPublicKey(),
	)

	if err = finalizer.FinalizeUsingObject(transacted); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.GetMetadataMutable().GetObjectDigest()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	privateKey := config.GetPrivateKey()

	if err = privateKey.Sign(
		transacted.GetMetadataMutable().GetObjectDigest(),
		transacted.GetMetadataMutable().GetObjectSigMutable(),
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
	transacted object,
) (err error) {
	if err = finalizer.FinalizeUsingObject(transacted); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if slices.Contains(
		[]string{ids.TypeInventoryListV1},
		transacted.GetMetadataMutable().GetType().String(),
	) {
		return err
	}

	if err = transacted.Verify(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
