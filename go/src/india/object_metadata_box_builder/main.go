package object_metadata_box_builder

import (
	"fmt"
	"sort"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/object_metadata"
)

type Builder string_format_writer.Box

func (builder *Builder) AddBlobDigestIfNecessary(
	digest interfaces.MarklId,
	funcAbbreviate ids.FuncAbbreviateString,
) {
	value := digest.String()

	if funcAbbreviate != nil {
		abbreviatedDigestString, err := funcAbbreviate(digest)
		if err != nil {
			panic(err)
		}

		if abbreviatedDigestString != "" {
			value = abbreviatedDigestString
		} else {
			ui.Todo("abbreviation func produced empty string")
		}
	}

	if value == "" {
		return
	}

	field := string_format_writer.Field{
		Value:      "@" + value,
		ColorType:  string_format_writer.ColorTypeHash,
		NoTruncate: true,
	}

	builder.Contents.Append(field)
}

func (builder *Builder) AddRepoPubKey(
	metadata object_metadata.MetadataMutable,
) {
	builder.addMarklIdIfNotNull(metadata.GetRepoPubKey())
}

func (builder *Builder) AddObjectSig(metadata object_metadata.MetadataMutable) {
	builder.addMarklId(metadata.GetObjectSig())
}

func (builder *Builder) AddMotherSigIfNecessary(
	metadata object_metadata.MetadataMutable,
) {
	builder.addMarklIdIfNotNull(metadata.GetMotherObjectSig())
}

func (builder *Builder) addMarklIdIfNotNull(id interfaces.MarklId) {
	if id.IsNull() {
		return
	}

	builder.addMarklId(id)
}

func (builder *Builder) addMarklId(id interfaces.MarklId) {
	builder.addMarklIdWithColorType(id, id.GetPurpose(), string_format_writer.ColorTypeHash)
}

func (builder *Builder) addMarklIdLockWithColorType(
	key string,
	value interfaces.MarklId,
	colorType string_format_writer.ColorType,
) {
	builder.addMarklIdWithColorType(value, key, colorType)
}

func (builder *Builder) addMarklIdWithColorType(
	value interfaces.MarklId,
	key string,
	colorType string_format_writer.ColorType,
) {
	if key == "" {
		panic(fmt.Sprintf("empty key for markl id: %q", value))
	}

	builder.Contents.Append(string_format_writer.Field{
		Key:        key,
		Separator:  '@',
		Value:      value.String(),
		NoTruncate: true,
		ColorType:  colorType,
	})
}

func (builder *Builder) AddError(err error) {
	var errorGroup errors.Group

	if errors.As(err, &errorGroup) {
		for _, err := range errorGroup {
			builder.Contents.Append(
				string_format_writer.Field{
					Key:        "error",
					Value:      err.Error(),
					ColorType:  string_format_writer.ColorTypeUserData,
					NoTruncate: true,
				},
			)
		}
	} else {
		builder.Contents.Append(
			string_format_writer.Field{
				Key:        "error",
				Value:      err.Error(),
				ColorType:  string_format_writer.ColorTypeUserData,
				NoTruncate: true,
			},
		)
	}
}

func (builder *Builder) AddTai(metadata object_metadata.MetadataMutable) {
	builder.Contents.Append(string_format_writer.Field{
		Value:     metadata.GetTai().String(),
		ColorType: string_format_writer.ColorTypeHash,
	})
}

func (builder *Builder) AddType(
	metadata object_metadata.MetadataMutable,
) {
	builder.Contents.Append(string_format_writer.Field{
		Value:     metadata.GetType().String(),
		ColorType: string_format_writer.ColorTypeType,
	})
}

func (builder *Builder) AddTypeAndLock(
	metadata object_metadata.MetadataMutable,
) {
	typeLock := metadata.GetTypeLock()

	if typeLock.GetValue().IsEmpty() {
		builder.AddType(metadata)
	} else {
		builder.addMarklIdLockWithColorType(
			typeLock.GetKey().String(),
			typeLock.GetValue(),
			string_format_writer.ColorTypeType,
		)
	}
}

func (builder *Builder) AddTags(metadata object_metadata.MetadataMutable) {
	tagCount := metadata.GetTags().Len()
	builder.Contents.Grow(tagCount)

	for tag := range metadata.AllTags() {
		builder.Contents.Append(string_format_writer.Field{
			Value: tag.String(),
		})
	}

	tags := builder.Contents[builder.Contents.Len()-tagCount:]

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Value < tags[j].Value
	})
}

func (builder *Builder) AddTagsAndLocks(metadata object_metadata.MetadataMutable) {
	panic(errors.Err405MethodNotAllowed)
}

func (builder *Builder) AddDescription(metadata object_metadata.MetadataMutable) {
	builder.Contents.Append(string_format_writer.Field{
		Value:     metadata.GetDescription().StringWithoutNewlines(),
		ColorType: string_format_writer.ColorTypeUserData,
	})
}
