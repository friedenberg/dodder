package object_metadata_box_builder

import (
	"fmt"
	"sort"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
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
	metadata *object_metadata.Metadata,
) {
	builder.addMarklIdIfNotNull(metadata.GetRepoPubKey())
}

func (builder *Builder) AddObjectSig(metadata *object_metadata.Metadata) {
	builder.addMarklId(metadata.GetObjectSig())
}

func (builder *Builder) AddMotherSigIfNecessary(
	metadata *object_metadata.Metadata,
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
	lock object_metadata.Lock,
	colorType string_format_writer.ColorType,
) {
	key := lock.Key
	id := lock.Id

	builder.addMarklIdWithColorType(id, key, colorType)
}

func (builder *Builder) addMarklIdWithColorType(
	id interfaces.MarklId,
	key string,
	colorType string_format_writer.ColorType,
) {
	if key == "" {
		panic(fmt.Sprintf("empty key for markl id: %q", id))
	}

	builder.Contents.Append(string_format_writer.Field{
		Key:        key,
		Separator:  '@',
		Value:      id.String(),
		NoTruncate: true,
		ColorType:  colorType,
	})
}

func (builder *Builder) AddError(err error) {
	var errorGroup errors.Group

	if errors.As(err, &errorGroup) {
		for _, e := range errorGroup {
			builder.Contents.Append(
				string_format_writer.Field{
					Key:        "error",
					Value:      e.Error(),
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

func (builder *Builder) AddTai(metadata *object_metadata.Metadata) {
	builder.Contents.Append(string_format_writer.Field{
		Value:     metadata.Tai.String(),
		ColorType: string_format_writer.ColorTypeHash,
	})
}

func (builder *Builder) AddType(metadata *object_metadata.Metadata) {
	// builder.addMarklIdLockWithColorType(
	// 	metadata.GetLockfile().GetTypeLock(),
	// 	string_format_writer.ColorTypeType,
	// )
	builder.Contents.Append(string_format_writer.Field{
		Value:     metadata.Type.String(),
		ColorType: string_format_writer.ColorTypeType,
	})
}

// TODO modify this method to not allocate an intermediate slice and instead
// append to box.Contents and sort within that slice
func (builder *Builder) AddTags(metadata *object_metadata.Metadata) {
	if metadata.Tags == nil {
		return
	}

	tags := make([]string_format_writer.Field, 0, metadata.Tags.Len())

	for tag := range metadata.Tags.AllPtr() {
		tags = append(
			tags,
			string_format_writer.Field{
				Value: tag.String(),
			},
		)
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Value < tags[j].Value
	})

	for _, tag := range tags {
		builder.Contents.Append(tag)
	}
}

func (builder *Builder) AddDescription(metadata *object_metadata.Metadata) {
	builder.Contents.Append(string_format_writer.Field{
		Value:     metadata.Description.StringWithoutNewlines(),
		ColorType: string_format_writer.ColorTypeUserData,
	})
}
