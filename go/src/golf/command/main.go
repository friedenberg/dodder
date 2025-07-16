package command

// TODO add description
type Command interface {
	Run(Request)
}

type Description struct {
	Short, Long string
}

type CommandWithDescription interface {
	GetDescription() Description
}

var commands = map[string]Command{}

func Commands() map[string]Command {
	return commands
}

func Register(name string, cmd Command) {
	if _, ok := commands[name]; ok {
		panic("command added more than once: " + name)
	}

	commands[name] = cmd
}
