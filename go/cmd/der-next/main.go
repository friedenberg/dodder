package main

import (
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/romeo/cmd"
)

func main() {
	store_version.VCurrent = store_version.VNext
	store_version.VNext = store_version.VNull
	cmd.Run("der")
}
