package box_format

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata_fmt"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func MakeBoxTransactedArchive(
	env env_ui.Env,
	options options_print.Options,
) *BoxTransacted {
	po := options.
		WithPrintShas(true).
		WithExcludeFields(true).
		WithDescriptionInBox(true)

	co := env.FormatColorOptionsOut()
	co.OffEntirely = true

	format := MakeBoxTransacted(
		co,
		po,
		env.StringFormatWriterFields(
			string_format_writer.CliFormatTruncationNone,
			co,
		),
		ids.Abbr{},
		nil,
		nil,
		nil,
	)

	format.isArchive = true

	return format
}

func MakeBoxTransacted(
	co string_format_writer.ColorOptions,
	options options_print.Options,
	boxStringEncoder interfaces.StringEncoderTo[string_format_writer.Box],
	abbr ids.Abbr,
	fsItemReadWriter sku.FSItemReadWriter,
	relativePath env_dir.RelativePath,
	headerWriter string_format_writer.HeaderWriter[*sku.Transacted],
) *BoxTransacted {
	return &BoxTransacted{
		optionsColor:     co,
		optionsPrint:     options,
		boxStringEncoder: boxStringEncoder,
		abbr:             abbr,
		fsItemReadWriter: fsItemReadWriter,
		relativePath:     relativePath,
		headerWriter:     headerWriter,
	}
}

type BoxTransacted struct {
	optionsColor string_format_writer.ColorOptions
	optionsPrint options_print.Options

	boxStringEncoder interfaces.StringEncoderTo[string_format_writer.Box]
	headerWriter     string_format_writer.HeaderWriter[*sku.Transacted]

	abbr             ids.Abbr
	fsItemReadWriter sku.FSItemReadWriter
	relativePath     env_dir.RelativePath

	isArchive bool
}

func (format *BoxTransacted) EncodeStringTo(
	object *sku.Transacted,
	writer interfaces.WriterAndStringWriter,
) (n int64, err error) {
	var box string_format_writer.Box

	// box.Header.RightAligned = true

	// if f.optionsPrint.PrintTime && !f.optionsPrint.PrintTai {
	// 	t := sk.GetTai()
	// 	box.Header.Value = t.Format(string_format_writer.StringFormatDateTime)
	// }

	if format.headerWriter != nil {
		if err = format.headerWriter.WriteBoxHeader(&box.Header, object); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	box.Contents = slices.Grow(box.Contents, 10)

	if err = format.addFieldsObjectIds(
		object,
		&box,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = format.addFieldsMetadata(
		format.optionsPrint,
		object,
		format.optionsPrint.DescriptionInBox,
		&box,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	b := &object.Metadata.Description

	if !format.optionsPrint.DescriptionInBox && !b.IsEmpty() {
		box.Trailer = append(
			box.Trailer,
			string_format_writer.Field{
				Value:              b.StringWithoutNewlines(),
				ColorType:          string_format_writer.ColorTypeUserData,
				DisableValueQuotes: true,
			},
		)
	}

	if n, err = format.boxStringEncoder.EncodeStringTo(box, writer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (f *BoxTransacted) makeFieldExternalObjectIdsIfNecessary(
	sk *sku.Transacted,
) (field string_format_writer.Field, err error) {
	field = string_format_writer.Field{
		ColorType: string_format_writer.ColorTypeId,
	}

	if !sk.ExternalObjectId.IsEmpty() {
		oid := &sk.ExternalObjectId
		// TODO quote as necessary
		field.Value = (&ids.ObjectIdStringerSansRepo{ObjectIdLike: oid}).String()
	}

	return
}

func (f *BoxTransacted) makeFieldObjectId(
	sk *sku.Transacted,
) (field string_format_writer.Field, empty bool, err error) {
	oid := &sk.ObjectId

	empty = oid.IsEmpty()

	oidString := (&ids.ObjectIdStringerSansRepo{oid}).String()

	if f.abbr.ZettelId.Abbreviate != nil &&
		oid.GetGenre() == genres.Zettel {
		if oidString, err = f.abbr.ZettelId.Abbreviate(oid); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	field = string_format_writer.Field{
		Value:              oidString,
		DisableValueQuotes: true,
		ColorType:          string_format_writer.ColorTypeId,
	}

	return
}

func (f *BoxTransacted) addFieldsObjectIds(
	sk *sku.Transacted,
	box *string_format_writer.Box,
) (err error) {
	var external string_format_writer.Field

	if external, err = f.makeFieldExternalObjectIdsIfNecessary(
		sk,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var internal string_format_writer.Field
	var externalEmpty bool

	if internal, externalEmpty, err = f.makeFieldObjectId(sk); err != nil {
		err = errors.Wrap(err)
		return
	}

	switch {
	// case internal.Value != "" && external.Value != "":
	// 	if strings.HasPrefix(external.Value, strings.TrimPrefix(internal.Value,
	// "!")) {
	// 		box.Contents = append(box.Contents, external)
	// 	} else {
	// 		box.Contents = append(box.Contents, internal, external)
	// 	}

	case externalEmpty && external.Value != "":
		box.Contents = append(box.Contents, external)

	case internal.Value != "":
		box.Contents = append(box.Contents, internal)

	case external.Value != "":
		box.Contents = append(box.Contents, external)

	default:
		err = errors.ErrorWithStackf("empty id")
		return
	}

	return
}

func (format *BoxTransacted) addFieldsMetadata(
	options options_print.Options,
	object *sku.Transacted,
	includeDescriptionInBox bool,
	box *string_format_writer.Box,
) (err error) {
	metadata := object.GetMetadata()

	if options.PrintShas &&
		(options.PrintEmptyShas || !metadata.Blob.IsNull()) {
		var shaString string

		if shaString, err = object_metadata_fmt.MetadataShaString(
			metadata,
			format.abbr.Sha.Abbreviate,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		box.Contents = append(
			box.Contents,
			object_metadata_fmt.MetadataFieldShaString(shaString),
		)
	}

	if options.PrintTai && object.GetGenre() != genres.InventoryList {
		box.Contents = append(
			box.Contents,
			object_metadata_fmt.MetadataFieldTai(metadata),
		)
	}

	if format.isArchive && !object.Metadata.RepoSig.IsEmpty() {
		box.Contents = append(
			box.Contents,
			object_metadata_fmt.MetadataFieldRepoPubKey(metadata),
			object_metadata_fmt.MetadataFieldRepoSig(metadata),
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

	if !options.ExcludeFields && !format.isArchive {
		box.Contents = append(box.Contents, metadata.Fields...)
	}

	return
}
