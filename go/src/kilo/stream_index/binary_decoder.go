package stream_index

import (
	"bytes"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/keys"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

var binaryFieldOrder = []keys.Binary{
	keys.Sigil,
	keys.ObjectId,
	keys.Blob,
	keys.RepoPubKey,
	keys.RepoSig,
	keys.Description,
	keys.Tag,
	keys.Tai,
	keys.Type,
	keys.DigestParentMetadataParentObjectId,
	keys.DigestMetadataParentObjectId,
	keys.DigestMetadataWithoutTai,
	keys.CacheParentTai,
	keys.CacheTagImplicit,
	keys.CacheTagExpanded,
	keys.CacheTags,
}

func makeBinary(s ids.Sigil) binaryDecoder {
	return binaryDecoder{
		queryGroup: sku.MakePrimitiveQueryGroupWithSigils(s),
		sigil:      s,
	}
}

func makeBinaryWithQueryGroup(
	query sku.PrimitiveQueryGroup,
	sigil ids.Sigil,
) binaryDecoder {
	if query == nil {
		query = sku.MakePrimitiveQueryGroup()
	}

	if !query.HasHidden() {
		sigil.Add(ids.SigilHidden)
	}

	return binaryDecoder{
		queryGroup: query,
		sigil:      sigil,
	}
}

type binaryDecoder struct {
	bytes.Buffer
	binaryField

	sigil         ids.Sigil
	queryGroup    sku.PrimitiveQueryGroup
	limitedReader io.LimitedReader
}

func (bf *binaryDecoder) readFormatExactly(
	r io.ReaderAt,
	sk *skuWithRangeAndSigil,
) (n int64, err error) {
	bf.binaryField.Reset()
	bf.Buffer.Reset()

	var n1 int
	var n2 int64

	b := make([]byte, sk.ContentLength)

	n1, err = r.ReadAt(b, sk.Offset)
	n += int64(n1)

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = errors.Wrap(err)
		return
	}

	buf := bytes.NewBuffer(b)

	n1, bf.ContentLength, err = ohio.ReadFixedUInt16(buf)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n2, err = bf.readSigil(sk, buf)
	n += n2

	if err != nil {
		err = errors.Wrapf(err, "ExpectedContentLenght: %d", bf.ContentLength)
		return
	}

	for buf.Len() > 0 {
		n2, err = bf.binaryField.ReadFrom(buf)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = bf.readFieldKey(sk.Transacted); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (decoder *binaryDecoder) readFormatAndMatchSigil(
	r io.Reader,
	sk *skuWithRangeAndSigil,
) (n int64, err error) {
	decoder.binaryField.Reset()
	decoder.Buffer.Reset()

	var n1 int
	var n2 int64

	// loop thru entries to find the next one that matches the current sigil
	// when found, break the loop and deserialize it and return
	for {
		n1, decoder.ContentLength, err = ohio.ReadFixedUInt16(r)
		n += int64(n1)

		if err != nil {
			if errors.IsEOF(err) {
				// no-op
				// TODO why might this ever be the case
			} else if errors.Is(err, io.ErrUnexpectedEOF) && n == 0 {
				err = io.EOF
			}

			err = errors.WrapExceptSentinel(err, io.EOF)

			return
		}

		contentLength64 := int64(decoder.ContentLength)

		decoder.limitedReader.R = r
		decoder.limitedReader.N = contentLength64

		n2, err = decoder.readSigil(sk, &decoder.limitedReader)
		n += n2

		if err != nil {
			err = errors.Wrapf(
				err,
				"ExpectedContentLenght: %d",
				contentLength64,
			)
			return
		}

		n2, err = decoder.binaryField.ReadFrom(&decoder.limitedReader)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = decoder.readFieldKey(sk.Transacted); err != nil {
			err = errors.Wrap(err)
			return
		}

		genre := genres.Must(sk.Transacted)
		query, ok := decoder.queryGroup.Get(genre)

		// TODO-D4 use query to decide whether to read and inflate or skip
		if ok {
			sigil := query.GetSigil()

			wantsHidden := sigil.IncludesHidden()
			wantsHistory := sigil.IncludesHistory()
			isLatest := sk.Contains(ids.SigilLatest)
			isHidden := sk.Contains(ids.SigilHidden)

			if (wantsHistory && wantsHidden) ||
				(wantsHidden && isLatest) ||
				(wantsHistory && !isHidden) ||
				(isLatest && !isHidden) {
				break
			}

			if query.ContainsObjectId(&sk.ObjectId) &&
				(sigil.ContainsOneOf(ids.SigilHistory) ||
					sk.ContainsOneOf(ids.SigilLatest)) {
				break
			}
		}

		// TODO-D4 replace with buffered seeker
		// discard the next record
		if _, err = io.Copy(io.Discard, &decoder.limitedReader); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	for decoder.limitedReader.N > 0 {
		n2, err = decoder.binaryField.ReadFrom(&decoder.limitedReader)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = decoder.readFieldKey(sk.Transacted); err != nil {
			err = errors.Wrapf(err, "Sku: %#v", sk.Transacted)
			return
		}
	}

	return
}

var errExpectedSigil = errors.New("expected sigil")

func (bf *binaryDecoder) readSigil(
	sk *skuWithRangeAndSigil,
	r io.Reader,
) (n int64, err error) {
	if n, err = bf.binaryField.ReadFrom(r); err != nil {
		err = errors.Wrapf(err, "Range: %q", sk.Range)
		return
	}

	if bf.Binary != keys.Sigil {
		err = errors.Wrapf(errExpectedSigil, "Key: %s", bf.Binary)
		return
	}

	if _, err = sk.Sigil.ReadFrom(&bf.Content); err != nil {
		err = errors.Wrap(err)
		return
	}

	sk.SetDormant(sk.IncludesHidden())

	return
}

func (bf *binaryDecoder) readFieldKey(
	object *sku.Transacted,
) (err error) {
	switch bf.Binary {
	case keys.Blob:
		if _, err = object.Metadata.Blob.ReadFrom(&bf.Content); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.RepoPubKey:
		if err = object.Metadata.GetPubKeyMutable().UnmarshalBinary(
			bf.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.RepoSig:
		if err = object.Metadata.GetObjectSigMutable().UnmarshalBinary(
			bf.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Description:
		if err = object.Metadata.Description.Set(bf.Content.String()); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Tag:
		var e ids.Tag

		if err = e.Set(bf.Content.String()); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = object.AddTagPtrFast(&e); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.ObjectId:
		if _, err = object.ObjectId.ReadFrom(&bf.Content); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Tai:
		if _, err = object.Metadata.Tai.ReadFrom(&bf.Content); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheParentTai:
		if _, err = object.Metadata.Cache.ParentTai.ReadFrom(&bf.Content); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.Type:
		if err = object.Metadata.Type.Set(bf.Content.String()); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestParentMetadataParentObjectId:
		if err = object.Metadata.GetMotherObjectDigestMutable().UnmarshalBinary(
			bf.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestMetadataParentObjectId:
		if err = object.Metadata.GetObjectDigestMutable().UnmarshalBinary(
			bf.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.DigestMetadataWithoutTai:
		if _, err = object.Metadata.SelfWithoutTai.ReadFrom(
			&bf.Content,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheTagImplicit:
		var e ids.Tag

		if err = e.Set(bf.Content.String()); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = object.Metadata.Cache.AddTagsImplicitPtr(&e); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheTagExpanded:
		var e ids.Tag

		if err = e.Set(bf.Content.String()); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = object.Metadata.Cache.AddTagExpandedPtr(&e); err != nil {
			err = errors.Wrap(err)
			return
		}

	case keys.CacheTags:
		var e tag_paths.PathWithType

		if _, err = e.ReadFrom(&bf.Content); err != nil {
			err = errors.WrapExceptSentinel(err, io.EOF)
			return
		}

		object.Metadata.Cache.TagPaths.AddPath(&e)

	default:
		err = errors.ErrorWithStackf("unsupported key: %s", bf.Binary)
		return
	}

	return
}
