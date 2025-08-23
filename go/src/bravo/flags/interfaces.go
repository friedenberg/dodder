package flags

// TODO modify this to expose a `GetCLIFlags() []string` method
type CommandComponentWriter interface {
	SetFlagSet(*FlagSet)
}
