package interfaces

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// If a Value has an IsBoolFlag() bool method returning true,
// the command-line parser makes -name equivalent to -name=true
// rather than using the next command-line argument.
//
// Set is called once, in command line order, for each flag present.
// The flag package may call the [String] method with a zero-valued receiver,
// such as a nil pointer.
type FlagValue interface {
	StringerSetter
}

type CLIFlagIOWrapper interface {
	FlagValue
	IOWrapper
}

// TODO rename to CLIFlagDefinitions
// TODO add completion options
type CLIFlagDefinitions interface {
	BoolVar(variable *bool, name string, defaultValue bool, usage string)
	StringVar(variable *string, name string, defaultValue string, usage string)
	Var(value FlagValue, name string, usage string)
	Func(name, usage string, funk func(string) error)
	IntVar(variable *int, name string, defaultValue int, usage string)
}

type CommandComponentWriter interface {
	SetFlagDefinitions(CLIFlagDefinitions)
}
