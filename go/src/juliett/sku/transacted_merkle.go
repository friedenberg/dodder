package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/object_fmt_digest"
	"code.linenisgreat.com/dodder/go/src/india/object_finalizer"
)

func (transacted *Transacted) SetMother(mother *Transacted) (err error) {
	motherSig := transacted.Metadata.GetMotherObjectSigMutable()

	if mother == nil {
		motherSig.Reset()
		return err
	}

	if err = motherSig.SetMarklId(
		markl.FormatIdEd25519Sig,
		mother.Metadata.GetObjectSig().GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = motherSig.SetPurpose(
		markl.PurposeObjectMotherSigV1,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) AssertObjectDigestAndObjectSigNotNull() (err error) {
	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectDigest()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectSig()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) Verify() (err error) {
	pubKey := transacted.Metadata.GetRepoPubKey()

	if err = markl.AssertIdIsNotNull(
		pubKey,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.Metadata.GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = pubKey.Verify(
		transacted.Metadata.GetObjectDigest(),
		transacted.Metadata.GetObjectSig(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

type funcCalcDigest func(object_fmt_digest.Format, object_fmt_digest.FormatterContext) (interfaces.MarklId, error)

type ObjectDigestWriteMap = object_finalizer.ObjectDigestWriteMap

func (transacted *Transacted) GetDigestWriteMapWithMerkle() ObjectDigestWriteMap {
	return ObjectDigestWriteMap{
		markl.PurposeV5MetadataDigestWithoutTai: &transacted.Metadata.SelfWithoutTai,
		markl.PurposeObjectDigestV1:             transacted.Metadata.GetObjectDigestMutable(),
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
	funcCalcDigest := object_fmt_digest.GetDigestForContext

	if debug {
		funcCalcDigest = object_fmt_digest.GetDigestForContextDebug
	}

	waitGroup := errors.MakeWaitGroupParallel()

	for formatId, id := range formats {
		var format object_fmt_digest.Format

		if format, err = object_fmt_digest.FormatForPurposeOrError(
			formatId,
		); err != nil {
			err = errors.Wrap(err)
			return err
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
		return err
	}

	return err
}

func (transacted *Transacted) MakeDigestCalcFunc(
	funcCalcDigest funcCalcDigest,
	format object_fmt_digest.Format,
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
	format object_fmt_digest.Format,
	digest interfaces.MutableMarklId,
) (err error) {
	var actual interfaces.MarklId

	if actual, err = funcCalcDigest(
		format,
		transacted,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = digest.SetPurpose(format.GetPurpose()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer markl.PutBlobId(actual)

	if err = digest.SetMarklId(
		actual.GetMarklFormat().GetMarklFormatId(),
		actual.GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
