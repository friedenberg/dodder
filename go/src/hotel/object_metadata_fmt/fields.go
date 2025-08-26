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

func MetadataBlobIdString(
	metadata *object_metadata.Metadata,
	abbr ids.FuncAbbreviateString,
) (digestString string, err error) {
	digest := sha.MustWithMerkleId(metadata.GetBlobDigest())
	defer sha.Env{}.PutBlobId(digest)

	digestString = digest.String()

	if abbr != nil {
		var abbreviatedDigestString string

		if abbreviatedDigestString, err = abbr(digest); err != nil {
			err = errors.Wrap(err)
			return
		}

		if abbreviatedDigestString != "" {
			digestString = abbreviatedDigestString
		} else {
			ui.Todo("abbreviate sha produced empty string")
		}
	}

	return
}

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

func MetadataFieldBlobIdString(
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
		Value:      metadata.GetRepoPubKey().String(),
		NoTruncate: true,
		ColorType:  string_format_writer.ColorTypeHash,
	}
}

func MetadataFieldRepoSig(
	metadata *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:      metadata.GetObjectSig().String(),
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
