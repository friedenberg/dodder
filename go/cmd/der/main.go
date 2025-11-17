package main

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/sierra/commands_dodder"
)

func main() {
	utility := commands_dodder.GetUtility("dodder")
	utility.Run(os.Args)
}
