package main

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/november/commands_madder"
)

func main() {
	utility := commands_madder.GetUtility()
	utility.Run(os.Args)
}
