package env_local

import (
	"code.linenisgreat.com/dodder/go/src/hotel/env_ui"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
)

type (
	ui  = env_ui.Env
	dir = env_dir.Env
)

type Env interface {
	ui
	dir
}

type env struct {
	ui
	dir
}

func Make(ui ui, dir dir) env {
	return env{
		ui:  ui,
		dir: dir,
	}
}
