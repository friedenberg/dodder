package stream_index

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"math"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/key_bytes"
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
	case key_bytes.Sigil:
		s := object.Sigil
		s.Add(encoder.Sigil)

		if object.Metadata.Cache.Dormant.Bool() {
			s.Add(ids.SigilHidden)
		}

		if n, err = encoder.writeFieldByteReader(s); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.Blob:
		if n, err = encoder.writeMerkleId(
			object.Metadata.GetBlobDigest(),
			true,
			encoder.Binary.String(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.RepoPubKey:
		if n, err = encoder.writeFieldBinaryMarshaler(
			object.Metadata.GetRepoPubKey(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.RepoSig:
		merkleId := object.Metadata.GetObjectSig()

		// TODO change to false once dropping V8
		if n, err = encoder.writeMerkleId(
			merkleId,
			true,
			encoder.Binary.String(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.Description:
		if object.Metadata.Description.IsEmpty() {
			return
		}

		if n, err = encoder.writeFieldBinaryMarshaler(&object.Metadata.Description); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.CacheParentTai:
		if n, err = encoder.writeFieldWriterTo(&object.Metadata.Cache.ParentTai); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.Tag:
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

	case key_bytes.ObjectId:
		if n, err = encoder.writeFieldWriterTo(&object.ObjectId); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.Tai:
		if n, err = encoder.writeFieldWriterTo(&object.Metadata.Tai); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.Type:
		if object.Metadata.Type.IsEmpty() {
			return
		}

		if n, err = encoder.writeFieldBinaryMarshaler(&object.Metadata.Type); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.SigParentMetadataParentObjectId:
		if n, err = encoder.writeFieldMerkleId(
			object.Metadata.GetMotherObjectSig(),
			true,
			encoder.Binary.String(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.DigestMetadataWithoutTai:
		if n, err = encoder.writeFieldMerkleId(
			&object.Metadata.SelfWithoutTai,
			true,
			encoder.Binary.String(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.DigestMetadataParentObjectId:
		if n, err = encoder.writeFieldMerkleId(
			object.Metadata.GetObjectDigest(),
			true,
			encoder.Binary.String(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case key_bytes.CacheTagImplicit:
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

	case key_bytes.CacheTagExpanded:
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

	case key_bytes.CacheTags:
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
	merkleId interfaces.MarklId,
	allowNull bool,
	key string,
) (n int64, err error) {
	if !allowNull {
		if err = markl.AssertIdIsNotNull(merkleId); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if n, err = encoder.writeFieldBinaryMarshaler(
		merkleId,
	); err != nil {
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
	merkleId interfaces.MarklId,
	allowNull bool,
	key string,
) (n int64, err error) {
	if merkleId.IsNull() {
		if !allowNull {
			err = markl.AssertIdIsNotNull(merkleId)
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
