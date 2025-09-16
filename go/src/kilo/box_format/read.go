package box_format

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/unicorn"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
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
			return
		}
	}

	if scanner.Error() != nil {
		err = errors.Wrap(scanner.Error())
		return
	}

	n = scanner.N()

	if format.optionsPrint.BoxDescriptionInBox {
		return
	}

	// TODO extract into dedicated parser and make incompatible with
	// BoxTransactedWithSignature
	if err = object.Metadata.Description.ReadFromBoxScanner(scanner); err != nil {
		err = errors.Wrap(err)
		return
	}

	n = scanner.N()

	return
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

		return
	}

	seq := scanner.GetSeq()

	if !seq.MatchAll(doddish.TokenMatcherOp(doddish.OpGroupOpen)) {
		err = ErrNotABox
		scanner.Unscan()

		return
	}

	if !scanner.ConsumeSpacesOrErrorOnFalse() {
		if scanner.Error() != nil {
			err = errors.Wrap(scanner.Error())
		} else {
			err = io.ErrUnexpectedEOF
		}

		return
	}

	return
}

// TODO switch to returning ErrBoxParse
func (format *BoxTransacted) readStringFormatBox(
	scanner *doddish.Scanner,
	object *sku.Transacted,
) (err error) {
	if err = format.openBox(scanner); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF, ErrNotABox)
		return
	}

	{
		if !scanner.ScanDotAllowedInIdentifiers() {
			if scanner.Error() != nil {
				err = errors.Wrap(scanner.Error())
			} else {
				err = io.ErrUnexpectedEOF
			}

			return
		}

		seq := scanner.GetSeq()

		if err = object.ObjectId.ReadFromSeq(seq); err != nil {
			err = nil
			object.ObjectId.Reset()

			if seq.MatchAll(doddish.TokenTypeLiteral) {
				if err = object.ExternalObjectId.Set(seq.String()); err != nil {
					err = errors.Wrap(err)
					return
				}
			} else if ok, left, _, _ := seq.PartitionFavoringLeft(
				doddish.TokenMatcherOp(doddish.OpPathSeparator),
			); ok && left.Len() == 0 {
				if err = object.ExternalObjectId.Set(seq.String()); err != nil {
					err = errors.Wrap(err)
					return
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
						return
					}
				}

				if err = object.ExternalObjectId.Set(seq.String()); err != nil {
					err = errors.Wrap(err)
					return
				}

			} else {
				err = errors.ErrorWithStackf("unsupported seq: %q", seq)
				return
			}
		}

		if object.ObjectId.GetGenre() == genres.InventoryList {
			if err = object.Metadata.Tai.Set(object.ObjectId.String()); err != nil {
				err = errors.Wrap(err)
				return
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
		// ] ' '
		case seq.MatchAll(doddish.TokenTypeOperator):
			r := rune(seq.At(0).Contents[0])

			switch {
			case r == ']':
				break LOOP_AFTER_OID

			case unicorn.IsSpace(r):
				continue
			}

			// "value"
		case seq.MatchAll(doddish.TokenTypeLiteral):
			if err = object.Metadata.Description.Set(
				seq.String(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			continue

			// @abcd
		case seq.MatchAll(doddish.TokenMatcherOp('@'), doddish.TokenTypeIdentifier):
			fallthrough

			// key@abcd
		case seq.MatchAll(doddish.TokenTypeIdentifier, doddish.TokenMatcherOp('@'), doddish.TokenTypeIdentifier):
			if err = format.parseMarklIdTag(object, seq); err != nil {
				err = errors.Wrap(err)
				return
			}

			continue

			// "value"
		case seq.MatchAll(
			doddish.TokenTypeLiteral,
		):
			if err = object.Metadata.Description.Set(seq.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

			continue

			// key=value key="value"
		case seq.MatchStart(
			doddish.TokenTypeIdentifier,
			doddish.TokenMatcherOp(doddish.OpExact),
		) || seq.MatchStart(
			doddish.TokenTypeIdentifier,
			doddish.TokenMatcherOp(doddish.OpExact),
			doddish.TokenTypeLiteral,
		):

			value := seq[2:]

			field := string_format_writer.Field{
				Key:   string(seq.At(0).Contents),
				Value: value.String(),
			}

			field.ColorType = string_format_writer.ColorTypeUserData
			object.Metadata.Fields = append(object.Metadata.Fields, field)

			continue
		}

		if err = objectId.ReadFromSeq(seq); err != nil {
			err = nil
			scanner.Unscan()
			return
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
			if err = object.Metadata.Tai.Set(objectId.String()); err != nil {
				err = errors.Wrap(err)
				return
			}

		case genres.Type:
			if err = object.Metadata.Type.TodoSetFromObjectId(&objectId); err != nil {
				err = errors.Wrap(err)
				return
			}

		case genres.Tag:
			var tag ids.Tag

			if err = tag.TodoSetFromObjectId(&objectId); err != nil {
				err = errors.Wrap(err)
				return
			}

			if tag.IsDodderTag() {
				// ignore
				// err = errors.Err405MethodNotAllowed.Errorf("tag: %q", tag)
				continue
			} else {
				if err = object.AddTagPtr(&tag); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

		default:
			err = genres.MakeErrUnsupportedGenre(objectId.GetGenre())
			err = errors.Wrapf(err, "Seq: %q", seq)
			return
		}

		objectId.Reset()
	}

	if scanner.Error() != nil {
		err = errors.Wrap(scanner.Error())
		return
	}

	return
}

// expects `seq` to include `@` as the first token
func (format *BoxTransacted) parseOldBlobIdTag(
	object *sku.Transacted,
	seq doddish.Seq,
) (err error) {
	if err = markl.SetHexBytes(
		"sha256",
		// merkle.HRPObjectBlobDigestSha256V1,
		object.Metadata.GetBlobDigestMutable(),
		seq.At(1).Contents,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

var dodderTagMerkleIdGetterTypeMapping = map[string]func(*object_metadata.Metadata) interfaces.MutableMarklId{
	"":                             (*object_metadata.Metadata).GetBlobDigestMutable,
	markl.PurposeRepoPubKeyV1:      (*object_metadata.Metadata).GetRepoPubKeyMutable,
	markl.PurposeObjectSigV0:       (*object_metadata.Metadata).GetObjectSigMutable,
	markl.PurposeObjectSigV1:       (*object_metadata.Metadata).GetObjectSigMutable,
	markl.PurposeObjectMotherSigV1: (*object_metadata.Metadata).GetMotherObjectSigMutable,
}

// expects `seq` to include `@` as the first token
func (format *BoxTransacted) parseMarklIdTag(
	object *sku.Transacted,
	seq doddish.Seq,
) (err error) {
	var marklFormatId string
	var value []byte

	if seq.Len() == 3 {
		marklFormatId = string(seq.At(0).Contents)
		value = seq.At(2).Contents
	} else {
		value = seq.At(1).Contents
	}

	if getMutableMerkleIdMethod, ok := dodderTagMerkleIdGetterTypeMapping[marklFormatId]; ok {
		id := getMutableMerkleIdMethod(&object.Metadata)

		if err = markl.SetMarklIdWithFormatBlech32(
			id,
			marklFormatId,
			string(value),
		); err != nil {
			err = errors.Wrapf(err, "Seq: %q", seq)
			return
		}
	} else {
		err = errors.Wrap(ErrUnsupportedDodderTag{tag: string(value)})
		return
	}

	return
}
