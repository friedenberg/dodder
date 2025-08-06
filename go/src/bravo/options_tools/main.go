package options_tools

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"github.com/google/shlex"
)

type Options struct {
	Merge []string `toml:"merge"`
}

func (options *Options) SetFlagSet(flagSet *flag.FlagSet) {
	flagSet.Func(
		"merge-tool",
		"utility to launch for merge conflict resolution",
		func(value string) (err error) {
			if options.Merge, err = shlex.Split(value); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	)
}
