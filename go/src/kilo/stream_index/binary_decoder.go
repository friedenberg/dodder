package stream_index

import (
	"bytes"
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/key_bytes"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

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
	// TODO unembed
	bytes.Buffer
	// TODO unembed
	binaryField

	sigil         ids.Sigil
	queryGroup    sku.PrimitiveQueryGroup
	limitedReader io.LimitedReader
}

func (decoder *binaryDecoder) readFormatExactly(
	readerAt io.ReaderAt,
	object *objectWithCursorAndSigil,
) (n int64, err error) {
	decoder.binaryField.Reset()
	decoder.Buffer.Reset()

	var n1 int
	var n2 int64

	// TODO pool this
	bites := make([]byte, object.ContentLength)

	n1, err = readerAt.ReadAt(bites, object.Offset)
	n += int64(n1)

	if err == io.EOF {
		err = nil
	} else if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	buffer := bytes.NewBuffer(bites)

	n1, decoder.ContentLength, err = ohio.ReadFixedUInt16(buffer)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n2, err = decoder.readSigil(object, buffer)
	n += n2

	if err != nil {
		ui.Debug().Printf("%T: %x", readerAt, buffer.Bytes())
		err = errors.Wrapf(
			err,
			"ExpectedContentLength: %d, ActualBytes: %d",
			decoder.ContentLength,
			buffer.Len(),
		)
		return n, err
	}

	for buffer.Len() > 0 {
		n2, err = decoder.binaryField.ReadFrom(buffer)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		if err = decoder.readFieldKey(object.Transacted); err != nil {
			err = errors.Wrap(err)
			return n, err
		}
	}

	return n, err
}

func (decoder *binaryDecoder) readFormatAndMatchSigil(
	r io.Reader,
	sk *objectWithCursorAndSigil,
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

			return n, err
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
			return n, err
		}

		n2, err = decoder.binaryField.ReadFrom(&decoder.limitedReader)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		if err = decoder.readFieldKey(sk.Transacted); err != nil {
			err = errors.Wrap(err)
			return n, err
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
			return n, err
		}
	}

	for decoder.limitedReader.N > 0 {
		n2, err = decoder.binaryField.ReadFrom(&decoder.limitedReader)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		if err = decoder.readFieldKey(sk.Transacted); err != nil {
			err = errors.Wrapf(err, "Sku: %#v", sk.Transacted)
			return n, err
		}
	}

	return n, err
}

var errExpectedSigil = errors.New("expected sigil")

func (decoder *binaryDecoder) readSigil(
	object *objectWithCursorAndSigil,
	reader io.Reader,
) (n int64, err error) {
	if n, err = decoder.binaryField.ReadFrom(reader); err != nil {
		err = errors.Wrapf(err, "Cursor: %q", object.Cursor)
		return n, err
	}

	if decoder.Key != key_bytes.Sigil {
		err = errors.Wrapf(
			errExpectedSigil,
			"Key: %s, ContentLength: %d, Reader: %T",
			decoder.Key,
			decoder.ContentLength,
			reader,
		)
		return n, err
	}

	if _, err = object.Sigil.ReadFrom(&decoder.Content); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	object.SetDormant(object.IncludesHidden())

	return n, err
}

func (decoder *binaryDecoder) readFieldKey(
	object *sku.Transacted,
) (err error) {
	switch decoder.Key {
	case key_bytes.Blob:
		if err = object.Metadata.GetBlobDigestMutable().UnmarshalBinary(
			decoder.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.RepoPubKey:
		if err = object.Metadata.GetRepoPubKeyMutable().UnmarshalBinary(
			decoder.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.RepoSig:
		if err = object.Metadata.GetObjectSigMutable().UnmarshalBinary(
			decoder.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.Description:
		if err = object.Metadata.Description.Set(decoder.Content.String()); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.Tag:
		var e ids.Tag

		if err = e.Set(decoder.Content.String()); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = object.AddTagPtrFast(&e); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.ObjectId:
		if _, err = object.ObjectId.ReadFrom(&decoder.Content); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.Tai:
		if _, err = object.Metadata.Tai.ReadFrom(&decoder.Content); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.CacheParentTai:
		if _, err = object.Metadata.Cache.ParentTai.ReadFrom(&decoder.Content); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.Type:
		if err = object.Metadata.Type.Set(decoder.Content.String()); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.SigParentMetadataParentObjectId:
		if err = object.Metadata.GetMotherObjectSigMutable().UnmarshalBinary(
			decoder.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.DigestMetadataParentObjectId:
		if err = object.Metadata.GetObjectDigestMutable().UnmarshalBinary(
			decoder.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.DigestMetadataWithoutTai:
		if err = object.Metadata.SelfWithoutTai.UnmarshalBinary(
			decoder.Content.Bytes(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.CacheTagImplicit:
		var tag ids.Tag

		if err = tag.Set(decoder.Content.String()); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = object.Metadata.Cache.AddTagsImplicitPtr(&tag); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.CacheTagExpanded:
		var tag ids.Tag

		if err = tag.Set(decoder.Content.String()); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = object.Metadata.Cache.AddTagExpandedPtr(&tag); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case key_bytes.CacheTags:
		var tag tag_paths.PathWithType

		if _, err = tag.ReadFrom(&decoder.Content); err != nil {
			err = errors.WrapExceptSentinel(err, io.EOF)
			return err
		}

		object.Metadata.Cache.TagPaths.AddPath(&tag)

	default:
		err = errors.ErrorWithStackf("unsupported key: %s", decoder.Key)
		return err
	}

	return err
}

func unmarshalMarklId(id interfaces.MutableMarklId, bites []byte) (err error) {
	unmarshaler := markl.IdBinaryDecodingFormatTypeData{
		MutableMarklId: id,
	}

	if err = unmarshaler.UnmarshalBinary(
		bites,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
