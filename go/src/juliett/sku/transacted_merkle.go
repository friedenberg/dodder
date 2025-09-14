package sku

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
)

func (transacted *Transacted) SetMother(mother *Transacted) (err error) {
	motherSig := transacted.Metadata.GetMotherObjectSigMutable()

	if mother == nil {
		motherSig.Reset()
		return
	}

	if err = motherSig.SetMarklId(
		markl.FormatIdSigEd25519,
		mother.Metadata.GetObjectSig().GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = motherSig.SetPurpose(
		markl.PurposeObjectMotherSigV1,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) AssertObjectDigestAndObjectSigNotNull() (err error) {
	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectDigest(),
		"object-digest",
	); err != nil {
		err = errors.Wrapf(err, "Object: %q", String(transacted))
		return
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectSig(),
		"object-sig",
	); err != nil {
		err = errors.Wrapf(err, "Object: %q", String(transacted))
		return
	}

	return
}

// calculates the object digests using the object's repo pubkey
func (transacted *Transacted) FinalizeUsingObject() (err error) {
	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetRepoPubKey(),
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return transacted.FinalizeUsingRepoPubKey(
		transacted.Metadata.GetRepoPubKey(),
	)
}

// calculates the object digests using the provided repo pubkey
func (transacted *Transacted) FinalizeUsingRepoPubKey(
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
			return
		}
	}

	if err = transacted.CalculateDigests(
		false,
		transacted.GetDigestWriteMapWithMerkle(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO remove / rename
func (transacted *Transacted) CalculateObjectDigests() (err error) {
	return transacted.CalculateDigests(
		false,
		transacted.GetDigestWriteMapWithoutMerkle(),
	)
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

	if transacted.Metadata.GetRepoPubKey().GetPurpose() == "" {
		panic("empty pbukey format")
	}

	return
}

func (transacted *Transacted) FinalizeAndSignOverwrite(
	config genesis_configs.ConfigPrivate,
) (err error) {
	// TODO populate format ids
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
	if err = markl.AssertIdIsNull(
		transacted.Metadata.GetRepoPubKey(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.AssertIdIsNull(
		transacted.Metadata.GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	transacted.Metadata.GetRepoPubKeyMutable().ResetWithMarklId(
		config.GetPublicKey(),
	)

	if err = transacted.FinalizeUsingObject(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectDigest(),
		"object-digest",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	privateKey := config.GetPrivateKey()

	if err = markl.Sign(
		privateKey,
		transacted.Metadata.GetObjectDigest(),
		config.GetObjectSigMarklTypeId(),
		markl.FormatIdSigEd25519,
		transacted.Metadata.GetObjectSigMutable(),
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

	if err = markl.AssertIdIsNotNull(
		pubKey,
		"repo-pubkey",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectDigest(),
		"object-digest",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectSig(),
		"object-sig",
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.Verify(
		pubKey,
		transacted.Metadata.GetObjectDigest(),
		transacted.Metadata.GetObjectSig(),
	); err != nil {
		err = errors.Wrapf(err, "Object: %q", String(transacted))
		return
	}

	return
}

type funcCalcDigest func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.MarklId, error)

type ObjectDigestWriteMap map[string]interfaces.MutableMarklId

func (transacted *Transacted) GetDigestWriteMapWithMerkle() ObjectDigestWriteMap {
	return ObjectDigestWriteMap{
		markl.PurposeV5MetadataDigestWithoutTai: &transacted.Metadata.SelfWithoutTai,
		markl.PurposeObjectDigestV1:       transacted.Metadata.GetObjectDigestMutable(),
	}
}

func (transacted *Transacted) GetDigestWriteMapWithoutMerkle() ObjectDigestWriteMap {
	return ObjectDigestWriteMap{
		markl.PurposeV5MetadataDigestWithoutTai: &transacted.Metadata.SelfWithoutTai,
	}
}

// calculates the respective digests
func (transacted *Transacted) CalculateDigests(
	debug bool,
	formats ObjectDigestWriteMap,
) (err error) {
	funcCalcDigest := object_inventory_format.GetDigestForContext

	if debug {
		funcCalcDigest = object_inventory_format.GetDigestForContextDebug
	}

	waitGroup := errors.MakeWaitGroupParallel()

	for formatId, id := range formats {
		var format object_inventory_format.Format

		if format, err = object_inventory_format.FormatForMarklFormatIdError(
			formatId,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		waitGroup.Do(
			transacted.MakeDigestCalcFunc(
				funcCalcDigest,
				format,
				id,
			),
		)
	}

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (transacted *Transacted) MakeDigestCalcFunc(
	funcCalcDigest funcCalcDigest,
	format object_inventory_format.Format,
	digest interfaces.MutableMarklId,
) errors.FuncErr {
	return func() (err error) {
		return transacted.CalculateDigest(
			funcCalcDigest,
			format,
			digest,
		)
	}
}

func (transacted *Transacted) CalculateDigest(
	funcCalcDigest funcCalcDigest,
	format object_inventory_format.Format,
	digest interfaces.MutableMarklId,
) (err error) {
	if err = digest.SetPurpose(format.GetMarklTypeId()); err != nil {
		err = errors.Wrap(err)
		return
	}

	var actual interfaces.MarklId

	if actual, err = funcCalcDigest(
		format,
		transacted,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer markl.PutBlobId(actual)

	if err = digest.SetMarklId(
		actual.GetMarklFormat().GetMarklFormatId(),
		actual.GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
