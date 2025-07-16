package commands

import (
	"bufio"
	"bytes"
	"flag"
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
)

func PrintUsage(ctx errors.Context, in error) {
	if in != nil {
		defer ctx.CancelWithError(in)
	}

	ui.Err().Print("Usage for dodder:")

	commands := command.Commands()

	flagSets := make([]*flag.FlagSet, 0, len(commands))

	for name, cmd := range commands {
		flagSet := flag.NewFlagSet(name, flag.ContinueOnError)

		if cmd, ok := cmd.(interfaces.CommandComponent); ok {
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

func PrintSubcommandUsage(flags flag.FlagSet) {
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
