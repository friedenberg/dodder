package command_components_dodder

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/flag_policy"
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/india/object_metadata"
)

type ObjectMetadata struct{}

func (cmd ObjectMetadata) GetFlagValueMetadataTags(
	metadata object_metadata.IMetadataMutable,
) interfaces.FlagValue {
	// TODO add support for tag_paths
	fes := flags.MakeWithPolicy(
		flag_policy.FlagPolicyAppend,
		func() string {
			return metadata.GetIndex().GetTagPaths().String()
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
	metadata object_metadata.IMetadataMutable,
) interfaces.FlagValue {
	return metadata.GetDescriptionMutable()
}

func (cmd ObjectMetadata) GetFlagValueMetadataType(
	metadata object_metadata.IMetadataMutable,
) interfaces.FlagValue {
	return metadata.GetTypePtr()
}
