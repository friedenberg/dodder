package box_format

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

// TODO make this error for invalid input
func (format *BoxTransacted) ReadStringFormat(
	object *sku.Transacted,
	runeScanner io.RuneScanner,
) (n int64, err error) {
	scanner := doddish.MakeScanner(runeScanner)

	if err = format.readStringFormatBox(scanner, object); err != nil {
		if err == ErrNotABox {
			err = nil
		} else {
			err = errors.WrapExceptSentinel(err, io.EOF, ErrBoxReadSeq{})
			return n, err
		}
	}

	if scanner.Error() != nil {
		err = errors.Wrap(scanner.Error())
		return n, err
	}

	n = scanner.N()

	if format.optionsPrint.BoxDescriptionInBox {
		return n, err
	}

	// TODO extract into dedicated parser and make incompatible with
	// BoxTransactedWithSignature
	if err = object.GetMetadataMutable().GetDescriptionMutable().ReadFromBoxScanner(
		scanner,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n = scanner.N()

	return n, err
}

func (format *BoxTransacted) openBox(
	scanner *doddish.Scanner,
) (err error) {
	if !scanner.ScanSkipSpace() {
		if scanner.Error() != nil {
			err = errors.Wrap(scanner.Error())
		} else {
			err = io.EOF
		}

		return err
	}

	seq := scanner.GetSeq()

	if !seq.MatchAll(doddish.TokenMatcherOp(doddish.OpGroupOpen)) {
		err = ErrNotABox
		scanner.Unscan()

		return err
	}

	if !scanner.ConsumeSpacesOrErrorOnFalse() {
		if scanner.Error() != nil {
			err = errors.Wrap(scanner.Error())
		} else {
			err = io.ErrUnexpectedEOF
		}

		return err
	}

	return err
}

// TODO switch to returning ErrBoxParse
func (format *BoxTransacted) readStringFormatBox(
	scanner *doddish.Scanner,
	object *sku.Transacted,
) (err error) {
	if err = format.openBox(scanner); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF, ErrNotABox)
		return err
	}

	{
		if !scanner.ScanDotAllowedInIdentifiers() {
			if scanner.Error() != nil {
				err = errors.Wrap(scanner.Error())
			} else {
				err = io.ErrUnexpectedEOF
			}

			return err
		}

		seq := scanner.GetSeq()

		if err = object.ObjectId.ReadFromSeq(seq); err != nil {
			err = nil
			object.ObjectId.Reset()

			if seq.MatchAll(doddish.TokenTypeLiteral) {
				if err = object.ExternalObjectId.Set(seq.String()); err != nil {
					err = errors.Wrap(err)
					return err
				}
			} else if ok, left, _, _ := seq.PartitionFavoringLeft(
				doddish.TokenMatcherOp(doddish.OpPathSeparator),
			); ok && left.Len() == 0 {
				if err = object.ExternalObjectId.Set(seq.String()); err != nil {
					err = errors.Wrap(err)
					return err
				}
			} else if ok, left, right := seq.MatchEnd(
				doddish.TokenMatcherOp(doddish.OpSigilExternal),
				doddish.TokenTypeIdentifier,
			); ok {
				var g genres.Genre

				// left: one/uno, right: .zettel
				if err = g.Set(right.At(1).String()); err != nil {
					err = nil
				} else {
					if err = object.ObjectId.SetWithGenre(left.String(), g); err != nil {
						object.ObjectId.Reset()
						err = errors.Wrap(err)
						return err
					}
				}

				if err = object.ExternalObjectId.Set(seq.String()); err != nil {
					err = errors.Wrap(err)
					return err
				}

			} else {
				err = errors.ErrorWithStackf("unsupported seq: %q", seq)
				return err
			}
		}

		if object.ObjectId.GetGenre() == genres.InventoryList {
			if err = object.GetMetadataMutable().GetTaiMutable().Set(
				object.ObjectId.String(),
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	var objectId ids.ObjectId

LOOP_AFTER_OID:

	for scanner.ScanDotAllowedInIdentifiers() {
		seq := scanner.GetSeq()

		// TODO convert this into a decision tree based on token type sequences
		// instead of a switch
		switch {
		case seq.MatchAll(doddish.TokenMatcherOp(']')):
			break LOOP_AFTER_OID

			// ' '
		case seq.MatchAll(doddish.TokenMatcherOp(' ')):
			continue

			// "value"
		case seq.MatchAll(doddish.TokenTypeLiteral):
			if err = object.GetMetadataMutable().GetDescriptionMutable().Set(
				seq.String(),
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

			continue

			// @digest
		case seq.MatchAll(doddish.TokenMatcherBlobDigest...):
			if err = format.parseMarklIdTag(object, seq); err != nil {
				err = errors.Wrap(err)
				return err
			}

			continue

			// !type
		case seq.MatchAll(doddish.TokenMatcherType...):
			if err = object.GetMetadataMutable().GetTypeMutable().Set(
				seq.String(),
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

			continue

			// !type@sig
		case seq.MatchAll(doddish.TokenMatcherTypeLock...):
			typeLock := object.GetMetadataMutable().GetTypeLockMutable()

			if err = typeLock.Key.Set(
				seq[0:2].String(),
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

			if err = markl.SetMarklIdWithFormatBlech32(
				&typeLock.Value,
				"",
				seq[2:].String(),
			); err != nil {
				err = errors.Wrapf(err, "Seq: %q", seq)
				return err
			}

			continue

			// key@abcd
		case seq.MatchAll(doddish.TokenMatcherDodderTag...):
			if err = format.parseMarklIdTag(object, seq); err != nil {
				err = errors.Wrap(err)
				return err
			}

			continue

			// key=value key="value"
		case seq.MatchStart(doddish.TokenMatcherKeyValue...) || seq.MatchStart(doddish.TokenMatcherKeyValueLiteral...):
			value := seq[2:]

			field := string_format_writer.Field{
				Key:   string(seq.At(0).Contents),
				Value: value.String(),
			}

			field.ColorType = string_format_writer.ColorTypeUserData
			object.GetMetadataMutable().GetIndexMutable().GetFieldsMutable().Append(field)

			continue
		}

		if err = objectId.ReadFromSeq(seq); err != nil {
			err = nil
			scanner.Unscan()
			return err
		}

		// if strings.Contains(objectId.String(), "dodder") {
		// 	objectIdString := objectId.String()
		// 	defer func() {
		// 		err = errors.Join(errors.Errorf("object contained dodder tag: %q",
		// objectIdString))
		// 	}()
		// }

		genre := objectId.GetGenre()

		switch genre {
		case genres.InventoryList:
			// TODO make more performant
			if err = object.GetMetadataMutable().GetTaiMutable().Set(
				objectId.String(),
			); err != nil {
				err = errors.Wrap(err)
				return err
			}

		case genres.Tag:
			var tag ids.Tag

			if err = tag.TodoSetFromObjectId(&objectId); err != nil {
				err = errors.Wrap(err)
				return err
			}

			if tag.IsDodderTag() {
				// ignore
				// err = errors.Err405MethodNotAllowed.Errorf("tag: %q", tag)
				continue
			} else {
				if err = object.AddTagPtr(&tag); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}

		default:
			err = genres.MakeErrUnsupportedGenre(objectId.GetGenre())
			err = errors.Wrapf(err, "Seq: %q", seq)
			return err
		}

		objectId.Reset()
	}

	if scanner.Error() != nil {
		err = errors.Wrap(scanner.Error())
		return err
	}

	return err
}

// expects `seq` to include `@` as the first token
func (format *BoxTransacted) parseMarklIdTag(
	object *sku.Transacted,
	seq doddish.Seq,
) (err error) {
	var purposeId string
	var value []byte

	if seq.Len() == 3 {
		purposeId = string(seq.At(0).Contents)
		value = seq.At(2).Contents
	} else {
		// blobs have only the `@` symbol and no purpose, so we implicitly decide
		// the purpose when parsing
		purposeId = markl.PurposeBlobDigestV1
		value = seq.At(1).Contents
	}

	metadata := object.GetMetadataMutable()

	var id interfaces.MutableMarklId

	switch markl.GetPurpose(purposeId).GetPurposeType() {
	case markl.PurposeTypeBlobDigest:
		id = metadata.GetBlobDigestMutable()

	case markl.PurposeTypeObjectMotherSig:
		id = metadata.GetMotherObjectSigMutable()

	case markl.PurposeTypeObjectSig:
		id = metadata.GetObjectSigMutable()

	case markl.PurposeTypeRepoPubKey:
		id = metadata.GetRepoPubKeyMutable()

	default:
		err = errors.Wrap(ErrUnsupportedDodderTag{tag: seq.String()})
		return err
	}

	if err = markl.SetMarklIdWithFormatBlech32(
		id,
		purposeId,
		string(value),
	); err != nil {
		err = errors.Wrapf(err, "Seq: %q", seq)
		return err
	}

	return err
}
