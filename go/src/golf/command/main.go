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

var commands = map[string]map[string]Command{
	"dodder": {},
	"madder": {},
}

func CommandsFor(util string) map[string]Command {
	return commands[util]
}

func Register(name string, cmd Command) {
	RegisterFor("dodder", name, cmd)
	RegisterFor("madder", name, cmd)
}

func RegisterFor(util string, name string, cmd Command) {
	if _, ok := commands[util][name]; ok {
		panic("command added more than once: " + name)
	}

	commands[util][name] = cmd
}
