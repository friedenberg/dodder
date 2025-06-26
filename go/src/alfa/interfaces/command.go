package interfaces

import "flag"

type CommandComponent interface {
	SetFlagSet(*flag.FlagSet)
}

type CommandLineIOWrapper interface {
	flag.Value
	IOWrapper
}
