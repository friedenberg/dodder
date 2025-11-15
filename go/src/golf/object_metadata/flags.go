package object_metadata

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/flag_policy"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
)

// TODO replace with command_components.ObjectMetadata
func (metadata *metadata) SetFlagDefinitions(flagSet interfaces.CLIFlagDefinitions) {
	metadata.SetFlagSetDescription(
		flagSet,
		"the description to use for created or updated Zettels",
	)

	metadata.SetFlagSetTags(
		flagSet,
		"the tags to use for created or updated object",
	)

	metadata.SetFlagSetType(
		flagSet,
		"the type for the created or updated object",
	)
}

func (metadata *metadata) SetFlagSetDescription(f interfaces.CLIFlagDefinitions, usage string) {
	f.Var(
		&metadata.Description,
		"description",
		usage,
	)
}

func (metadata *metadata) SetFlagSetTags(f interfaces.CLIFlagDefinitions, usage string) {
	// TODO add support for tag_paths
	fes := flags.MakeWithPolicy(
		flag_policy.FlagPolicyAppend,
		func() string {
			return metadata.Cache.TagPaths.String()
		},
		func(value string) (err error) {
			values := strings.SplitSeq(value, ",")

			for tagString := range values {
				if err = metadata.AddTagString(tagString); err != nil {
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

	f.Var(
		fes,
		"tags",
		usage,
	)
}

func (metadata *metadata) SetFlagSetType(f interfaces.CLIFlagDefinitions, usage string) {
	f.Func(
		"type",
		usage,
		func(v string) (err error) {
			return metadata.Type.Set(v)
		},
	)
}
