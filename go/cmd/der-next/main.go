package main

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/quebec/commands"
)

func main() {
	store_version.VCurrent = store_version.VNext
	store_version.VNext = store_version.VNull
	utility := commands.GetUtility("dodder")
	utility.Run(os.Args)
}
