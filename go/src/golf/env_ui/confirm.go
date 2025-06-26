package env_ui

import (
	"fmt"
)

func (env *env) Confirm(message string) (success bool) {
	if !env.GetIn().IsTty() {
		env.GetErr().Print(
			"stdin is not a tty, unable to get permission to continue",
		)

		return
	}

	var err error

	env.GetErr().Printf(
		"%s (y/*)",
		message,
	)

	var answer rune
	var n int

	if n, err = fmt.Fscanf(env.GetInFile(), "%c\n", &answer); err != nil {
		env.GetErr().Printf("failed to read answer: %s", err)
		return
	}

	if n != 1 {
		env.GetErr().Printf("failed to read at exactly 1 answer: %s", err)
		return
	}

	if answer == 'y' || answer == 'Y' {
		success = true
	}

	return
}
