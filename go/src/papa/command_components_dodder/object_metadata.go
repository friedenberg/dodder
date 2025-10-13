package command_components_dodder

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/flag_policy"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flag"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

type ObjectMetadata struct{}

func (cmd ObjectMetadata) GetFlagValueMetadataTags(
	metadata *object_metadata.Metadata,
) interfaces.FlagValue {
	// TODO add support for tag_paths
	fes := flag.Make(
		flag_policy.FlagPolicyAppend,
		func() string {
			return metadata.Cache.TagPaths.String()
		},
		func(v string) (err error) {
			vs := strings.Split(v, ",")

			for _, v := range vs {
				if err = metadata.AddTagString(v); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}

			return err
		},
		func() {
			metadata.ResetTags()
		},
	)

	return fes
}

func (cmd ObjectMetadata) GetFlagValueMetadataDescription(
	metadata *object_metadata.Metadata,
) interfaces.FlagValue {
	return &metadata.Description
}

func (cmd ObjectMetadata) GetFlagValueMetadataType(
	metadata *object_metadata.Metadata,
) interfaces.FlagValue {
	return &metadata.Type
}
