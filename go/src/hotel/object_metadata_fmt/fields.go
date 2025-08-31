package object_metadata_fmt

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func MetadataFieldError(
	err error,
) []string_format_writer.Field {
	var errorGroup errors.Group

	if errors.As(err, &errorGroup) {
		out := make([]string_format_writer.Field, 0, errorGroup.Len())

		for _, e := range errorGroup {
			out = append(
				out,
				string_format_writer.Field{
					Key:        "error",
					Value:      e.Error(),
					ColorType:  string_format_writer.ColorTypeUserData,
					NoTruncate: true,
				},
			)
		}

		return out
	} else {
		return []string_format_writer.Field{
			{
				Key:        "error",
				Value:      err.Error(),
				ColorType:  string_format_writer.ColorTypeUserData,
				NoTruncate: true,
			},
		}
	}
}

func MetadataFieldTai(
	metadata *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     metadata.Tai.String(),
		ColorType: string_format_writer.ColorTypeHash,
	}
}

func MetadataFieldType(
	metadata *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     metadata.Type.String(),
		ColorType: string_format_writer.ColorTypeType,
	}
}

func MetadataFieldTags(
	metadata *object_metadata.Metadata,
) []string_format_writer.Field {
	if metadata.Tags == nil {
		return nil
	}

	tags := make([]string_format_writer.Field, 0, metadata.Tags.Len())

	for t := range metadata.Tags.AllPtr() {
		tags = append(
			tags,
			string_format_writer.Field{
				Value: t.String(),
			},
		)
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Value < tags[j].Value
	})

	return tags
}

func MetadataFieldDescription(
	metadata *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     metadata.Description.StringWithoutNewlines(),
		ColorType: string_format_writer.ColorTypeUserData,
	}
}
