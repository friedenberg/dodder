package stream_index

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"math"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/keys"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type binaryEncoder struct {
	bytes.Buffer
	binaryField
	ids.Sigil
}

func (encoder *binaryEncoder) updateSigil(
	writerAt io.WriterAt,
	sigil ids.Sigil,
	offset int64,
) (err error) {
	sigil.Add(encoder.Sigil)
	// 2 uint8 + offset + 2 uint8 + Schlussel
	offset = int64(2) + offset + int64(3)

	var n int

	if n, err = writerAt.WriteAt([]byte{sigil.Byte()}, offset); err != nil {
		err = errors.Wrap(err)
		return
	}

	if n != 1 {
		err = catgut.MakeErrLength(1, int64(n), nil)
		return
	}

	return
}

func (encoder *binaryEncoder) writeFormat(
	writer io.Writer,
	object skuWithSigil,
) (n int64, err error) {
	encoder.Buffer.Reset()

	for _, f := range binaryFieldOrder {
		encoder.binaryField.Reset()
		encoder.Binary = f

		if _, err = encoder.writeFieldKey(object); err != nil {
			err = errors.Wrapf(
				err,
				"Key: %q, Sku: %s",
				encoder.Binary,
				sku.String(object.Transacted),
			)
			return
		}
	}

	encoder.binaryField.Reset()
	rawContentLength := encoder.Len()

	if rawContentLength > math.MaxUint16 {
		err = errContentLengthTooLarge
		return
	}

	encoder.ContentLength = uint16(rawContentLength)

	var n1 int
	n1, err = ohio.WriteFixedUInt16(writer, encoder.ContentLength)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	var n2 int64
	n2, err = encoder.Buffer.WriteTo(writer)
	n += n2

	return
}

func (encoder *binaryEncoder) writeFieldKey(
	object skuWithSigil,
) (n int64, err error) {
	switch encoder.Binary {
	case keys.Sigil:
		s := object.Sigil
		s.Add(encoder.Sigil)

		if object.Metadata.Cache.Dormant.Bool() {
			s.Add(ids.SigilHidden)
		}

		if n, err = encoder.writeFieldByteReader(s); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Blob:
		if n, err = encoder.writeSha(&object.Metadata.Blob, true); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.RepoPubKey:
		if n, err = encoder.writeFieldBinaryMarshaler(
			object.Metadata.GetRepoPubKey(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.RepoSig:
		merkleId := object.Metadata.GetObjectSig()

		if n, err = encoder.writeMerkleId(
			merkleId,
			false,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Description:
		if object.Metadata.Description.IsEmpty() {
			return
		}

		if n, err = encoder.writeFieldBinaryMarshaler(&object.Metadata.Description); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheParentTai:
		if n, err = encoder.writeFieldWriterTo(&object.Metadata.Cache.ParentTai); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Tag:
		es := object.GetTags()

		for _, e := range quiter.SortedValues(es) {
			if e.IsVirtual() {
				continue
			}

			if e.String() == "" {
				err = errors.ErrorWithStackf("empty tag in %q", es)
				return
			}

			var n1 int64
			n1, err = encoder.writeFieldBinaryMarshaler(&e)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case keys.ObjectId:
		if n, err = encoder.writeFieldWriterTo(&object.ObjectId); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Tai:
		if n, err = encoder.writeFieldWriterTo(&object.Metadata.Tai); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Type:
		if object.Metadata.Type.IsEmpty() {
			return
		}

		if n, err = encoder.writeFieldBinaryMarshaler(&object.Metadata.Type); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestParentMetadataParentObjectId:
		if n, err = encoder.writeFieldMerkleId(object.Metadata.GetMotherObjectDigest(), true); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestMetadataWithoutTai:
		if n, err = encoder.writeSha(&object.Metadata.SelfWithoutTai, true); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestMetadataParentObjectId:
		if n, err = encoder.writeFieldMerkleId(object.Metadata.GetObjectDigest(), false); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheTagImplicit:
		tags := object.Metadata.Cache.GetImplicitTags()

		for _, tag := range quiter.SortedValues(tags) {
			var n1 int64
			n1, err = encoder.writeFieldBinaryMarshaler(&tag)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case keys.CacheTagExpanded:
		tags := object.Metadata.Cache.GetExpandedTags()

		for _, tag := range quiter.SortedValues(tags) {
			var n1 int64
			n1, err = encoder.writeFieldBinaryMarshaler(&tag)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case keys.CacheTags:
		tags := object.Metadata.Cache.TagPaths

		for _, tag := range tags.Paths {
			var n1 int64
			n1, err = encoder.writeFieldWriterTo(tag)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	default:
		panic(fmt.Sprintf("unsupported key: %s", encoder.Binary))
	}

	return
}

func (encoder *binaryEncoder) writeMerkleId(
	merkleId interfaces.MerkleId,
	allowNull bool,
) (n int64, err error) {
	if err = merkle_ids.MakeErrIsNull(merkleId); err != nil {
		err = errors.Wrap(err)
		return
	}

	if n, err = encoder.writeFieldBinaryMarshaler(
		merkleId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO change to writeDigest
func (encoder *binaryEncoder) writeSha(
	sh *sha.Sha,
	allowNull bool,
) (n int64, err error) {
	if sh.IsNull() {
		if !allowNull {
			err = merkle_ids.MakeErrIsNull(sh)
		}

		return
	}

	if n, err = encoder.writeFieldWriterTo(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (encoder *binaryEncoder) writeFieldWriterTo(
	wt io.WriterTo,
) (n int64, err error) {
	_, err = wt.WriteTo(&encoder.Content)
	if err != nil {
		return
	}

	if n, err = encoder.binaryField.WriteTo(&encoder.Buffer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (encoder *binaryEncoder) writeFieldMerkleId(
	merkleId interfaces.MerkleId,
	allowNull bool,
) (n int64, err error) {
	if merkleId.IsNull() {
		if !allowNull {
			err = merkle_ids.MakeErrIsNull(merkleId)
		}

		return
	}

	if n, err = encoder.writeFieldBinaryMarshaler(merkleId); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (encoder *binaryEncoder) writeFieldBinaryMarshaler(
	binaryMarshaler encoding.BinaryMarshaler,
) (n int64, err error) {
	var bites []byte

	bites, err = binaryMarshaler.MarshalBinary()
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = ohio.WriteAllOrDieTrying(&encoder.Content, bites); err != nil {
		err = errors.WrapExceptSentinelAsNil(err, io.EOF)
		return
	}

	if n, err = encoder.binaryField.WriteTo(&encoder.Buffer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (encoder *binaryEncoder) writeFieldByteReader(
	byteReader io.ByteReader,
) (n int64, err error) {
	var bite byte

	bite, err = byteReader.ReadByte()
	if err != nil {
		return
	}

	err = encoder.Content.WriteByte(bite)
	if err != nil {
		return
	}

	if n, err = encoder.binaryField.WriteTo(&encoder.Buffer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
