package options_tools

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"github.com/google/shlex"
)

type Options struct {
	Merge []string `toml:"merge"`
}

var _ interfaces.CommandComponentWriter = (*Options)(nil)

func (options *Options) SetFlagDefinitions(flagSet interfaces.CommandLineFlagDefinitions) {
	flagSet.Func(
		"merge-tool",
		"utility to launch for merge conflict resolution",
		func(value string) (err error) {
			if options.Merge, err = shlex.Split(value); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		},
	)
}
