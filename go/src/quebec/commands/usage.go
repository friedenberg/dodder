package commands

import (
	"bufio"
	"bytes"
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
)

func PrintUsage(ctx interfaces.Context, err error) {
	if err != nil {
		defer ctx.Cancel(err)
	}

	ui.Err().Print("Usage for dodder:")

	commands := command.Commands()

	flagSets := make([]*flags.FlagSet, 0, len(commands))

	for name, cmd := range commands {
		flagSet := flags.NewFlagSet(name, flags.ContinueOnError)

		if cmd, ok := cmd.(flags.CommandComponentWriter); ok {
			cmd.SetFlagSet(flagSet)
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
