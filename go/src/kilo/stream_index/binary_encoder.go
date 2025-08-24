package stream_index

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"math"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/keys"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type binaryEncoder struct {
	bytes.Buffer
	binaryField
	ids.Sigil
}

func (bf *binaryEncoder) updateSigil(
	wa io.WriterAt,
	s ids.Sigil,
	offset int64,
) (err error) {
	s.Add(bf.Sigil)
	// 2 uint8 + offset + 2 uint8 + Schlussel
	offset = int64(2) + offset + int64(3)

	var n int

	if n, err = wa.WriteAt([]byte{s.Byte()}, offset); err != nil {
		err = errors.Wrap(err)
		return
	}

	if n != 1 {
		err = catgut.MakeErrLength(1, int64(n), nil)
		return
	}

	return
}

func (bf *binaryEncoder) writeFormat(
	w io.Writer,
	sk skuWithSigil,
) (n int64, err error) {
	bf.Buffer.Reset()

	for _, f := range binaryFieldOrder {
		bf.binaryField.Reset()
		bf.Binary = f

		if _, err = bf.writeFieldKey(sk); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sk)
			return
		}
	}

	bf.binaryField.Reset()
	rawContentLength := bf.Len()

	if rawContentLength > math.MaxUint16 {
		err = errContentLengthTooLarge
		return
	}

	bf.ContentLength = uint16(rawContentLength)

	var n1 int
	n1, err = ohio.WriteFixedUInt16(w, bf.ContentLength)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	var n2 int64
	n2, err = bf.Buffer.WriteTo(w)
	n += n2

	return
}

func (bf *binaryEncoder) writeFieldKey(
	sk skuWithSigil,
) (n int64, err error) {
	switch bf.Binary {
	case keys.Sigil:
		s := sk.Sigil
		s.Add(bf.Sigil)

		if sk.Metadata.Cache.Dormant.Bool() {
			s.Add(ids.SigilHidden)
		}

		if n, err = bf.writeFieldByteReader(s); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Blob:
		if n, err = bf.writeSha(&sk.Metadata.Blob, true); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.RepoPubKey:
		if n, err = bf.writeFieldBinaryMarshaler(
			&sk.Metadata.RepoPubkey,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.RepoSig:
		if n, err = bf.writeFieldBinaryMarshaler(
			&sk.Metadata.RepoSig,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Description:
		if sk.Metadata.Description.IsEmpty() {
			return
		}

		if n, err = bf.writeFieldBinaryMarshaler(&sk.Metadata.Description); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheParentTai:
		if n, err = bf.writeFieldWriterTo(&sk.Metadata.Cache.ParentTai); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Tag:
		es := sk.GetTags()

		for _, e := range quiter.SortedValues(es) {
			if e.IsVirtual() {
				continue
			}

			if e.String() == "" {
				err = errors.ErrorWithStackf("empty tag in %q", es)
				return
			}

			var n1 int64
			n1, err = bf.writeFieldBinaryMarshaler(&e)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case keys.ObjectId:
		if n, err = bf.writeFieldWriterTo(&sk.ObjectId); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Tai:
		if n, err = bf.writeFieldWriterTo(&sk.Metadata.Tai); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Type:
		if sk.Metadata.Type.IsEmpty() {
			return
		}

		if n, err = bf.writeFieldBinaryMarshaler(&sk.Metadata.Type); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestParentMetadataParentObjectId:
		if n, err = bf.writeSha(sk.Metadata.GetMotherDigest(), true); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestMetadataWithoutTai:
		if n, err = bf.writeSha(&sk.Metadata.SelfWithoutTai, true); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestMetadataParentObjectId:
		if n, err = bf.writeSha(sk.Metadata.GetDigest(), false); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheTagImplicit:
		tags := sk.Metadata.Cache.GetImplicitTags()

		for _, tag := range quiter.SortedValues(tags) {
			var n1 int64
			n1, err = bf.writeFieldBinaryMarshaler(&tag)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case keys.CacheTagExpanded:
		es := sk.Metadata.Cache.GetExpandedTags()

		for _, e := range quiter.SortedValues[ids.Tag](es) {
			var n1 int64
			n1, err = bf.writeFieldBinaryMarshaler(&e)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case keys.CacheTags:
		es := sk.Metadata.Cache.TagPaths

		for _, e := range es.Paths {
			var n1 int64
			n1, err = bf.writeFieldWriterTo(e)
			n += n1

			if err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	default:
		panic(fmt.Sprintf("unsupported key: %s", bf.Binary))
	}

	return
}

// TODO change to writeDigest
func (bf *binaryEncoder) writeSha(
	sh *sha.Sha,
	allowNull bool,
) (n int64, err error) {
	if sh.IsNull() {
		if !allowNull {
			err = blob_ids.MakeErrIsNull(sh)
		}

		return
	}

	if n, err = bf.writeFieldWriterTo(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (bf *binaryEncoder) writeFieldWriterTo(
	wt io.WriterTo,
) (n int64, err error) {
	_, err = wt.WriteTo(&bf.Content)
	if err != nil {
		return
	}

	if n, err = bf.binaryField.WriteTo(&bf.Buffer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (bf *binaryEncoder) writeFieldBinaryMarshaler(
	bm encoding.BinaryMarshaler,
) (n int64, err error) {
	var b []byte

	b, err = bm.MarshalBinary()
	if err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = ohio.WriteAllOrDieTrying(&bf.Content, b); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err != nil {
		err = errors.WrapExceptSentinelAsNil(err, io.EOF)
		return
	}

	if n, err = bf.binaryField.WriteTo(&bf.Buffer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (bf *binaryEncoder) writeFieldByteReader(
	br io.ByteReader,
) (n int64, err error) {
	var b byte

	b, err = br.ReadByte()
	if err != nil {
		return
	}

	err = bf.Content.WriteByte(b)
	if err != nil {
		return
	}

	if n, err = bf.binaryField.WriteTo(&bf.Buffer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
