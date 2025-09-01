package box_format

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata_fmt"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func MakeBoxCheckedOut(
	co string_format_writer.ColorOptions,
	options options_print.Options,
	boxWriter interfaces.StringEncoderTo[string_format_writer.Box],
	abbr ids.Abbr,
	fsItemReadWriter sku.FSItemReadWriter,
	relativePath env_dir.RelativePath,
	headerWriter string_format_writer.HeaderWriter[*sku.CheckedOut],
) *BoxCheckedOut {
	return &BoxCheckedOut{
		headerWriter: headerWriter,
		BoxTransacted: BoxTransacted{
			optionsColor:     co,
			optionsPrint:     options,
			boxStringEncoder: boxWriter,
			abbr:             abbr,
			fsItemReadWriter: fsItemReadWriter,
			relativePath:     relativePath,
		},
	}
}

type BoxCheckedOut struct {
	headerWriter string_format_writer.HeaderWriter[*sku.CheckedOut]
	BoxTransacted
}

func (format *BoxCheckedOut) EncodeStringTo(
	co *sku.CheckedOut,
	sw interfaces.WriterAndStringWriter,
) (n int64, err error) {
	var box string_format_writer.Box

	if format.headerWriter != nil {
		if err = format.headerWriter.WriteBoxHeader(&box.Header, co); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	box.Contents = slices.Grow(box.Contents, 10)

	var fds *sku.FSItem
	var errFS error

	external := co.GetSkuExternal()

	if format.fsItemReadWriter != nil {
		fds, errFS = format.fsItemReadWriter.ReadFSItemFromExternal(external)
	}

	if format.fsItemReadWriter == nil || errFS != nil ||
		!external.RepoId.IsEmpty() {
		if err = format.addFieldsExternalWithFSItem(
			external,
			&box,
			format.optionsPrint.BoxDescriptionInBox,
			fds,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = format.addFieldsFS(co, &box, fds); err != nil {
			err = errors.Wrapf(err, "CheckedOut: %s", co)
			return
		}
	}

	b := &external.Metadata.Description

	if !format.optionsPrint.BoxDescriptionInBox && !b.IsEmpty() {
		box.Trailer = append(
			box.Trailer,
			string_format_writer.Field{
				Value:              b.StringWithoutNewlines(),
				ColorType:          string_format_writer.ColorTypeUserData,
				DisableValueQuotes: true,
			},
		)
	}

	if n, err = format.boxStringEncoder.EncodeStringTo(box, sw); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (format *BoxCheckedOut) addFieldsExternalWithFSItem(
	external *sku.Transacted,
	box *string_format_writer.Box,
	includeDescriptionInBox bool,
	item *sku.FSItem,
) (err error) {
	if err = format.addFieldsObjectIdsWithFSItem(
		external,
		box,
		true,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = format.addFieldsMetadataWithFSItem(
		format.optionsPrint,
		external,
		includeDescriptionInBox,
		box,
		item,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// func (f *BoxTransacted) makeFieldExternalObjectIdsIfNecessary(
// 	sk *sku.Transacted,
// 	item *sku.FSItem,
// ) (field string_format_writer.Field, err error) {
// 	field = string_format_writer.Field{
// 		ColorType: string_format_writer.ColorTypeId,
// 	}

// 	switch {
// 	case (item == nil || item.Len() == 0) && !sk.ExternalObjectId.IsEmpty():
// 		oid := &sk.ExternalObjectId
// 		// TODO quote as necessary
// 		field.Value = (&ids.ObjectIdStringerSansRepo{oid}).String()

// 	case item != nil:
// 		switch item.GetCheckoutMode() {
// 		case checkout_mode.MetadataOnly, checkout_mode.MetadataAndBlob:
// 			// TODO quote as necessary
// 			if !item.Object.IsStdin() {
// 				field.Value = f.relativePath.Rel(item.Object.String())
// 			}
// 		}
// 	}

// 	return
// }

func (format *BoxCheckedOut) makeFieldObjectId(
	sk *sku.Transacted,
) (field string_format_writer.Field, empty bool, err error) {
	oid := &sk.ObjectId

	empty = oid.IsEmpty()

	oidString := (&ids.ObjectIdStringerSansRepo{ObjectIdLike: oid}).String()

	if format.abbr.ZettelId.Abbreviate != nil &&
		oid.GetGenre() == genres.Zettel {
		if oidString, err = format.abbr.ZettelId.Abbreviate(oid); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	field = string_format_writer.Field{
		Value:     oidString,
		ColorType: string_format_writer.ColorTypeId,
	}

	return
}

func (format *BoxTransacted) addFieldsObjectIdsWithFSItem(
	object *sku.Transacted,
	box *string_format_writer.Box,
	preferExternal bool,
) (err error) {
	var external string_format_writer.Field

	if external, err = format.makeFieldExternalObjectIdsIfNecessary(
		object,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var internal string_format_writer.Field
	var internalEmpty bool

	if internal, internalEmpty, err = format.makeFieldObjectId(object); err != nil {
		err = errors.Wrap(err)
		return
	}

	switch {
	case (internalEmpty || preferExternal) && external.Value != "":
		box.Contents = append(box.Contents, external)

	case internal.Value != "":
		box.Contents = append(box.Contents, internal)

	case external.Value != "":
		box.Contents = append(box.Contents, external)

	default:
		err = errors.Errorf("empty id")
		return
	}

	return
}

func (format *BoxCheckedOut) addFieldsMetadataWithFSItem(
	options options_print.Options,
	object *sku.Transacted,
	includeDescriptionInBox bool,
	box *string_format_writer.Box,
	item *sku.FSItem,
) (err error) {
	metadata := object.GetMetadata()

	if options.PrintBlobDigests &&
		(options.BoxPrintEmptyBlobIds || !metadata.GetBlobDigest().IsNull()) {
		box.Contents = object_metadata_fmt.AddBlobDigestIfNecessary(
			box.Contents,
			metadata.GetBlobDigest(),
			format.abbr.BlobId.Abbreviate,
		)
	}

	if options.BoxPrintTai && object.GetGenre() != genres.InventoryList {
		box.Contents = append(
			box.Contents,
			object_metadata_fmt.MetadataFieldTai(metadata),
		)
	}

	if !metadata.Type.IsEmpty() {
		box.Contents = append(
			box.Contents,
			object_metadata_fmt.MetadataFieldType(metadata),
		)
	}

	b := metadata.Description

	if includeDescriptionInBox && !b.IsEmpty() {
		box.Contents = append(
			box.Contents,
			object_metadata_fmt.MetadataFieldDescription(metadata),
		)
	}

	box.Contents = append(
		box.Contents,
		object_metadata_fmt.MetadataFieldTags(metadata)...,
	)

	if !options.BoxExcludeFields &&
		(item == nil || item.FDs.Len() == 0) {
		box.Contents = append(box.Contents, metadata.Fields...)
	}

	return
}

func (format *BoxCheckedOut) addFieldsFS(
	co *sku.CheckedOut,
	box *string_format_writer.Box,
	item *sku.FSItem,
) (err error) {
	m := item.GetCheckoutMode()

	var fdAlreadyWritten *fd.FD

	op := format.optionsPrint
	op.BoxExcludeFields = true

	switch co.GetState() {
	case checked_out_state.Unknown:
		err = errors.ErrorWithStackf("invalid state unknown")
		return

	case checked_out_state.Untracked:
		if err = format.addFieldsUntracked(co, box, item, op); err != nil {
			err = errors.Wrap(err)
			return
		}

		return

	case checked_out_state.Recognized:
		if err = format.addFieldsRecognized(co, box, item, op); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}

	var id string_format_writer.Field

	switch {
	case m == checkout_mode.BlobOnly || m == checkout_mode.BlobRecognized:
		id.Value = (&ids.ObjectIdStringerSansRepo{&co.GetSkuExternal().ObjectId}).String()

	case m.IncludesMetadata():
		id.Value = format.relativePath.Rel(item.Object.GetPath())
		fdAlreadyWritten = &item.Object

	default:
		if id, _, err = format.makeFieldObjectId(
			co.GetSkuExternal(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	id.ColorType = string_format_writer.ColorTypeId
	box.Contents = append(box.Contents, id)

	if co.GetState() == checked_out_state.Conflicted {
		// TODO handle conflicted state
	} else {
		if err = format.addFieldsMetadata(
			op,
			co.GetSkuExternal(),
			op.BoxDescriptionInBox,
			box,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if m == checkout_mode.BlobRecognized ||
			(m != checkout_mode.MetadataOnly && m != checkout_mode.None) {
			if err = format.addFieldsFSBlobExcept(
				item,
				fdAlreadyWritten,
				box,
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	return
}

func (format *BoxTransacted) addFieldsFSBlobExcept(
	fds *sku.FSItem,
	except *fd.FD,
	box *string_format_writer.Box,
) (err error) {
	if fds.FDs == nil {
		err = errors.ErrorWithStackf("FDSet.MutableSetLike was nil")
		return
	}

	for fd := range fds.FDs.All() {
		if except != nil && fd.Equals(except) {
			continue
		}

		if err = format.addFieldFSBlob(fd, box); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (format *BoxTransacted) addFieldFSBlob(
	fd *fd.FD,
	box *string_format_writer.Box,
) (err error) {
	box.Contents = append(
		box.Contents,
		string_format_writer.Field{
			Value:        format.relativePath.Rel(fd.GetPath()),
			ColorType:    string_format_writer.ColorTypeId,
			NeedsNewline: true,
		},
	)

	return
}

func (format *BoxTransacted) addFieldsUntracked(
	checkedOut *sku.CheckedOut,
	box *string_format_writer.Box,
	item *sku.FSItem,
	op options_print.Options,
) (err error) {
	// fdToPrint := &item.Blob

	// if co.GetSkuExternal().GetGenre() != genres.Zettel &&
	// !item.Object.IsEmpty() {
	// 	fdToPrint = &item.Object
	// }

	// box.Contents = append(
	// 	box.Contents,
	// 	string_format_writer.Field{
	// 		ColorType: string_format_writer.ColorTypeId,
	// 		Value:     f.relativePath.Rel(fdToPrint.GetPath()),
	// 	},
	// )
	if err = format.addFieldsObjectIdsWithFSItem(
		checkedOut.GetSkuExternal(),
		box,
		true,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = format.addFieldsMetadata(
		op,
		checkedOut.GetSkuExternal(),
		format.optionsPrint.BoxDescriptionInBox,
		box,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (format *BoxTransacted) addFieldsRecognized(
	co *sku.CheckedOut,
	box *string_format_writer.Box,
	item *sku.FSItem,
	op options_print.Options,
) (err error) {
	var id string_format_writer.Field

	if id, _, err = format.makeFieldObjectId(
		co.GetSkuExternal(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	id.ColorType = string_format_writer.ColorTypeId
	box.Contents = append(box.Contents, id)

	if err = format.addFieldsMetadata(
		op,
		co.GetSkuExternal(),
		op.BoxDescriptionInBox,
		box,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = format.addFieldsFSBlobExcept(
		item,
		nil,
		box,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
