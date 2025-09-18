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

var utilsToCommands = map[string]map[string]Command{
	"dodder": {},
	"madder": {},
}

func CommandsFor(util string) map[string]Command {
	switch util {
	case "der":
		util = "dodder"
	}

	return utilsToCommands[util]
}

// TODO switch to registering on default command
func Register(name string, cmd Command) {
	RegisterFor("dodder", name, cmd)
	RegisterFor("madder", name, cmd)
}

func RegisterFor(util string, name string, cmd Command) {
	if _, ok := utilsToCommands[util][name]; ok {
		panic("command added more than once: " + name)
	}

	utilsToCommands[util][name] = cmd
}
