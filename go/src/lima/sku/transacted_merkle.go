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

type (
	ObjectDigestWriteMap          = interfaces.DigestWriteMap
	ObjectDigestPurposeMarklIdSeq = interfaces.Seq2[string, interfaces.MutableMarklId]
)

func (transacted *Transacted) GetDigestWriteMapWithMerkle(
	defaultMarklFormatId string,
) ObjectDigestWriteMap {
	return ObjectDigestWriteMap{
		markl.PurposeV5MetadataDigestWithoutTai: transacted.GetMetadataMutable().GetSelfWithoutTaiMutable(),
		defaultMarklFormatId:                    transacted.GetMetadataMutable().GetObjectDigestMutable(),
	}
}

func (transacted *Transacted) CalculateDigests(
	formats ObjectDigestPurposeMarklIdSeq,
) (err error) {
	return transacted.calculateDigestsAndMaybeDebug(false, formats)
}

func (transacted *Transacted) CalculateDigestsDebug(
	formats ObjectDigestPurposeMarklIdSeq,
) (err error) {
	return transacted.calculateDigestsAndMaybeDebug(true, formats)
}

func (transacted *Transacted) calculateDigestsAndMaybeDebug(
	debug bool,
	formats ObjectDigestPurposeMarklIdSeq,
) (err error) {
	waitGroup := errors.MakeWaitGroupParallel()

	for formatId, id := range formats {
		waitGroup.Do(
			transacted.MakeDigestCalcFunc(
				formatId,
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
	purposeId string,
	digest interfaces.MutableMarklId,
) errors.FuncErr {
	return func() (err error) {
		return transacted.CalculateDigest2(
			purposeId,
			digest,
		)
	}
}

func (transacted *Transacted) CalculateDigest2(
	purposeId string,
	digest interfaces.MutableMarklId,
) (err error) {
	if err = object_fmt_digest.WriteDigest(
		purposeId,
		transacted,
		digest,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
