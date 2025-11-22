package object_finalizer

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/india/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

type (
	FinalizerGetter interface {
		GetObjectFinalizer() Finalizer
	}

	Finalizer = finalizer

	finalizer struct {
		index sku.IndexPrimitives
		// pubKey interfaces.MarklId
	}

	object interface {
		object_metadata.GetterMutable

		CalculateDigests(
			debug bool,
			formats interfaces.DigestWriteMap,
		) (err error)

		GetDigestWriteMapWithMerkle() interfaces.DigestWriteMap
		GetDigestWriteMapWithoutMerkle() interfaces.DigestWriteMap
		Verify() (err error)
	}
)

func Make() Finalizer {
	return finalizer{}
}

// func Make(index sku.IndexPrimitives) Finalizer {
// 	return finalizer{index: index}
// }

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
	object object,
	pubKey interfaces.MarklId,
) (err error) {
	metadataMutable := object.GetMetadataMutable()
	// TODO migrate this to config
	pubKeyMutable := metadataMutable.GetRepoPubKeyMutable()

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

	if err = finalizer.WriteLockfileIfNecessary(object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = object.CalculateDigests(
		false,
		object.GetDigestWriteMapWithMerkle(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (finalizer finalizer) WriteLockfileIfNecessary(
	object object,
) (err error) {
	if finalizer.index == nil {
		return err
	}

	return finalizer.WriteLockfile(object, finalizer.index.ReadOneMarklId)
}

func (finalizer finalizer) WriteLockfile(
	object object,
	funcReadOne sku.FuncReadOne,
) (err error) {
	metadata := object.GetMetadataMutable()

	if err = finalizer.writeTypeLockIfNecessary(
		metadata,
		metadata.GetType(),
		funcReadOne,
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
