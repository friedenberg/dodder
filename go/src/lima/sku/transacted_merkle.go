package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/kilo/object_fmt_digest"
)

func (transacted *Transacted) SetMother(mother *Transacted) (err error) {
	motherSig := transacted.GetMetadataMutable().GetMotherObjectSigMutable()

	if mother == nil {
		motherSig.Reset()
		return err
	}

	if err = motherSig.SetMarklId(
		markl.FormatIdEd25519Sig,
		mother.GetMetadata().GetObjectSig().GetBytes(),
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
		transacted.GetMetadata().GetObjectDigest()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = markl.AssertIdIsNotNull(
		transacted.GetMetadata().GetObjectSig()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (transacted *Transacted) Verify() (err error) {
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

type funcCalcDigest func(object_fmt_digest.Format, object_fmt_digest.FormatterContext) (interfaces.MarklId, error)

type ObjectDigestWriteMap = interfaces.DigestWriteMap

func (transacted *Transacted) GetDigestWriteMapWithMerkle(
	defaultMarklFormatId string,
) ObjectDigestWriteMap {
	return ObjectDigestWriteMap{
		markl.PurposeV5MetadataDigestWithoutTai: transacted.GetMetadataMutable().GetSelfWithoutTaiMutable(),
		defaultMarklFormatId:                    transacted.GetMetadataMutable().GetObjectDigestMutable(),
	}
}

func (transacted *Transacted) GetDigestWriteMapWithoutMerkle() ObjectDigestWriteMap {
	return ObjectDigestWriteMap{
		markl.PurposeV5MetadataDigestWithoutTai: transacted.GetMetadataMutable().GetSelfWithoutTaiMutable(),
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
