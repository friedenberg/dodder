package command

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sort"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
)

type Utility struct {
	name string
	cmds map[string]Cmd
}

func MakeUtility(name string) Utility {
	utility := Utility{
		name: name,
		cmds: make(map[string]Cmd),
	}

	return utility
}

func (utility Utility) GetName() string {
	return utility.name
}

func (utility Utility) GetCmd(name string) (Cmd, bool) {
	cmd, ok := utility.cmds[name]
	return cmd, ok
}

func (utility Utility) LenCmds() int {
	return len(utility.cmds)
}

func (utility Utility) AllCmds() interfaces.Seq2[string, Cmd] {
	return func(yield func(string, Cmd) bool) {
		for name, cmd := range utility.cmds {
			if !yield(name, cmd) {
				return
			}
		}
	}
}

func (utility Utility) AddCmd(name string, cmd Cmd) {
	if _, ok := utility.cmds[name]; ok {
		panic("subcommand added more than once: " + name)
	}

	utility.cmds[name] = cmd
}

func (utility Utility) MergeUtilityWithPrefix(
	otherUtility Utility,
	prefix string,
) Utility {
	for name, subcommand := range otherUtility.AllCmds() {
		if prefix != "" {
			name = fmt.Sprintf("%s-%s", prefix, name)
		}

		utility.AddCmd(name, subcommand)
	}

	return utility
}

func (utility Utility) MergeUtility(otherUtility Utility) Utility {
	return utility.MergeUtilityWithPrefix(otherUtility, "")
}

func (utility Utility) PrintUsage(ctx interfaces.Context, err error) {
	if err != nil {
		defer ctx.Cancel(err)
	}

	ui.Err().Print("Usage for dodder:")

	flagSets := make([]*flags.FlagSet, 0, utility.LenCmds())

	for name, cmd := range utility.AllCmds() {
		flagSet := flags.NewFlagSet(name, flags.ContinueOnError)

		if cmd, ok := cmd.(interfaces.CommandComponentWriter); ok {
			cmd.SetFlagDefinitions(flagSet)
		}

		flagSets = append(flagSets, flagSet)
	}

	sort.Slice(flagSets, func(i, j int) bool {
		return flagSets[i].Name() < flagSets[j].Name()
	})

	for _, f := range flagSets {
		ui.Err().Print(f.Name())
	}
}

// TODO evaluate need
func PrintSubcommandUsage(flags flags.FlagSet) {
	printTabbed := func(s string) {
		ui.Err().Print(s)
	}

	var b bytes.Buffer
	flags.SetOutput(&b)

	printTabbed(flags.Name())

	flags.PrintDefaults()

	scanner := bufio.NewScanner(&b)

	for scanner.Scan() {
		printTabbed(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		ui.Err().Print(err)
	}
}

func (utility Utility) Run(
	args []string,
) {
	utilityNameWithExtension := extendNameIfNecessary(utility.GetName())
	ctx := errors.MakeContextDefault()

	ctx.SetCancelOnSignals(
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
	)

	if err := ctx.Run(
		func(ctx interfaces.Context) {
			if len(args) <= 1 {
				utility.PrintUsage(
					ctx,
					errors.BadRequestf("No subcommand provided."),
				)

				return
			}

			var cmd Cmd
			var ok bool

			name := args[1]

			// TODO switch to context
			if cmd, ok = utility.GetCmd(name); !ok {
				utility.PrintUsage(
					ctx,
					errors.BadRequestf("No subcommand %q", name),
				)

				return
			}

			flagSet := flags.NewFlagSet(name, flags.ContinueOnError)

			if cmd, ok := cmd.(interfaces.CommandComponentWriter); ok {
				cmd.SetFlagDefinitions(flagSet)
			}

			args = args[2:]

			configCli := repo_config_cli.Default()
			configCli.SetFlagDefinitions(flagSet)

			if err := flagSet.Parse(args); err != nil {
				ctx.Cancel(err)
			}

			req := Request{
				Utility: utility,
				Context: ctx,
				Config:  configCli,
				FlagSet: flagSet,
				Args: &Args{
					Context: ctx,
					args:    flagSet.Args(),
				},
			}

			cmd.Run(req)
		},
	); err != nil {
		os.Exit(handleMainErrors(ctx, utilityNameWithExtension, err))
	}
}
