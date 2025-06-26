package object_metadata_fmt

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/bech32"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

func MetadataShaString(
	m *object_metadata.Metadata,
	abbr ids.FuncAbbreviateString,
) (v string, err error) {
	s := &m.Blob
	v = s.String()

	if abbr != nil {
		var v1 string

		sh := sha.Make(s)

		if v1, err = abbr(sh); err != nil {
			err = errors.Wrap(err)
			return
		}

		if v1 != "" {
			v = v1
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
	v string,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     "@" + v,
		ColorType: string_format_writer.ColorTypeHash,
	}
}

func MetadataFieldTai(
	m *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     m.Tai.String(),
		ColorType: string_format_writer.ColorTypeHash,
	}
}

func MetadataFieldRepoPubKey(
	m *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Key: "repo-pubkey",
		Value: bech32.Value{
			HRP:  repo_signing.HRPRepoPubKeyV1,
			Data: m.RepoPubKey,
		}.String(),
		NoTruncate: true,
		ColorType:  string_format_writer.ColorTypeHash,
	}
}

func MetadataFieldRepoSig(
	m *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Key: "repo-sig",
		Value: bech32.Value{
			HRP:  repo_signing.HRPRepoSigV1,
			Data: m.RepoSig,
		}.String(),
		NoTruncate: true,
		ColorType:  string_format_writer.ColorTypeHash,
	}
}

func MetadataFieldType(
	m *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     m.Type.String(),
		ColorType: string_format_writer.ColorTypeType,
	}
}

func MetadataFieldTags(
	m *object_metadata.Metadata,
) []string_format_writer.Field {
	if m.Tags == nil {
		return nil
	}

	out := make([]string_format_writer.Field, 0, m.Tags.Len())

	m.Tags.EachPtr(
		func(t *ids.Tag) (err error) {
			out = append(
				out,
				string_format_writer.Field{
					Value: t.String(),
				},
			)
			return
		},
	)

	sort.Slice(out, func(i, j int) bool {
		return out[i].Value < out[j].Value
	})

	return out
}

func MetadataFieldDescription(
	m *object_metadata.Metadata,
) string_format_writer.Field {
	return string_format_writer.Field{
		Value:     m.Description.StringWithoutNewlines(),
		ColorType: string_format_writer.ColorTypeUserData,
	}
}
