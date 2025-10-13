package main

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/quebec/commands"
)

func main() {
	utility := commands.GetUtility("dodder")
	utility.Run(os.Args)
}
