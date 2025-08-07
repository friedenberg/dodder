package interfaces

import "flag"

// TODO modify this to expose a `GetCLIFlags() []string` method
type CommandComponentWriter interface {
	SetFlagSet(*flag.FlagSet)
}

type CommandLineIOWrapper interface {
	flag.Value
	IOWrapper
}
