package main

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/sierra/commands_dodder"
)

func main() {
	store_version.VCurrent = store_version.VNext
	store_version.VNext = store_version.VNull
	utility := commands_dodder.GetUtility("dodder")
	utility.Run(os.Args)
}
