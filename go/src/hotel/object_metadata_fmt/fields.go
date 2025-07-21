package object_metadata_fmt

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func MetadataShaString(
	metadata *object_metadata.Metadata,
	abbr ids.FuncAbbreviateString,
) (value string, err error) {
	s := &metadata.Blob
	value = s.String()

	if abbr != nil {
		var v1 string

		sh := sha.MustWithDigester(s)

		if v1, err = abbr(sh); err != nil {
			err = errors.Wrap(err)
			return
		}

		if v1 != "" {
			value = v1
		} else {
			ui.Todo("abbreviate sha produced empty string")
		}
	}

	return
}

func MetadataFieldError(
	err error,
) []string_format_writer.Field {
	var me errors.Multi

	if errors.As(err, &me) {
		out := make([]string_format_writer.Field, 0, me.Len())

		for _, e := range me.Errors() {
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

func MetadataFieldShaString(
	value string,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     "@" + value,
		ColorType: string_format_writer.ColorTypeHash,
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

func MetadataFieldRepoPubKey(
	metadata *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:      metadata.GetRepoPubkeyValue().String(),
		NoTruncate: true,
		ColorType:  string_format_writer.ColorTypeHash,
	}
}

func MetadataFieldRepoSig(
	metadata *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:      metadata.GetRepoSigValue().String(),
		NoTruncate: true,
		ColorType:  string_format_writer.ColorTypeHash,
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

	metadata.Tags.EachPtr(
		func(t *ids.Tag) (err error) {
			tags = append(
				tags,
				string_format_writer.Field{
					Value: t.String(),
				},
			)
			return
		},
	)

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
