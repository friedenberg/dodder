package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
)

type funcCalcDigest func(object_inventory_format.Format, object_inventory_format.FormatterContext) (interfaces.MarklId, error)

type digestWriteMap map[string]interfaces.MutableMarklId

func (transacted *Transacted) getDigestWriteMapWithMerkle() digestWriteMap {
	return digestWriteMap{
		markl.FormatIdV5MetadataDigestWithoutTai: &transacted.Metadata.SelfWithoutTai,
		markl.FormatIdObjectDigestSha256V1:       transacted.Metadata.GetObjectDigestMutable(),
	}
}

func (transacted *Transacted) getDigestWriteMapWithoutMerkle() digestWriteMap {
	return digestWriteMap{
		markl.FormatIdV5MetadataDigestWithoutTai: &transacted.Metadata.SelfWithoutTai,
	}
}

// calculates the respective digests
func (transacted *Transacted) finalize(
	debug bool,
	formats digestWriteMap,
) (err error) {
	funcCalcDigest := object_inventory_format.GetDigestForContext

	if debug {
		funcCalcDigest = object_inventory_format.GetDigestForContextDebug
	}

	waitGroup := errors.MakeWaitGroupParallel()

	for formatId, id := range formats {
		waitGroup.Do(
			transacted.makeDigestCalcFunc(
				funcCalcDigest,
				formatId,
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

func (transacted *Transacted) makeDigestCalcFunc(
	funcCalcDigest funcCalcDigest,
	formatTypeId string,
	digest interfaces.MutableMarklId,
) errors.FuncErr {
	return func() (err error) {
		if err = digest.SetFormat(formatTypeId); err != nil {
			err = errors.Wrap(err)
			return
		}

		var objectFormat object_inventory_format.Format

		if objectFormat, err = object_inventory_format.FormatForMarklFormatIdError(
			digest.GetFormat(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

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
			actual.GetMarklType().GetMarklTypeId(),
			actual.GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (transacted *Transacted) CalculateObjectDigestSelfWithoutTai(
	funcCalcDigest funcCalcDigest,
) (err error) {
	if err = transacted.makeDigestCalcFunc(
		funcCalcDigest,
		markl.FormatIdV5MetadataDigestWithoutTai,
		&transacted.Metadata.SelfWithoutTai,
	)(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
