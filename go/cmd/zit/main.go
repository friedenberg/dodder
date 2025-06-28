package main

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/romeo/cmd"
)

func main() {
	os.Setenv(env_dir.EnvXDGUtilityNameOverride, "zit")
	cmd.Run("zit")
}
