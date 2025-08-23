package object_metadata

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/flag_policy"
	"code.linenisgreat.com/dodder/go/src/bravo/flag"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
)

// TODO replace with command_components.ObjectMetadata
func (metadata *Metadata) SetFlagSet(flagSet *flags.FlagSet) {
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

func (metadata *Metadata) SetFlagSetDescription(f *flags.FlagSet, usage string) {
	f.Var(
		&metadata.Description,
		"description",
		usage,
	)
}

func (metadata *Metadata) SetFlagSetTags(f *flags.FlagSet, usage string) {
	// TODO add support for tag_paths
	fes := flag.Make(
		flag_policy.FlagPolicyAppend,
		func() string {
			return metadata.Cache.TagPaths.String()
		},
		func(value string) (err error) {
			values := strings.SplitSeq(value, ",")

			for tagString := range values {
				if err = metadata.AddTagString(tagString); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			return
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

func (metadata *Metadata) SetFlagSetType(f *flags.FlagSet, usage string) {
	f.Func(
		"type",
		usage,
		func(v string) (err error) {
			return metadata.Type.Set(v)
		},
	)
}
