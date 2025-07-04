package env_repo

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

// Config used to initialize a repo for the first time
type BigBang struct {
	ids.Type
	Config *config_immutable.LatestPrivate

	Yin                  string
	Yang                 string
	ExcludeDefaultType   bool
	ExcludeDefaultConfig bool
	OverrideXDGWithCwd   bool
}

func (bigBang *BigBang) SetFlagSet(f *flag.FlagSet) {
	f.BoolVar(&bigBang.OverrideXDGWithCwd, "override-xdg-with-cwd", false, "")
	f.StringVar(&bigBang.Yin, "yin", "", "File containing list of zettel id left parts")
	f.StringVar(&bigBang.Yang, "yang", "", "File containing list of zettel id right parts")

	bigBang.Type = builtin_types.GetOrPanic(builtin_types.ImmutableConfigV1).Type
	bigBang.Config = config_immutable.Default()
	bigBang.Config.SetFlagSet(f)
}
