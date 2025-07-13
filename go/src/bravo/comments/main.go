package comments

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func Change(_ string) {
	ui.TodoP1("start logging this")
}

func Decide(_ string) {
	ui.TodoP1("start logging this")
}

func Refactor(_ string) {
	ui.TodoP1("start logging this")
}

func Parallelize() {
	ui.TodoP1("start logging this")
}

func Optimize() {
	ui.TodoP1("start logging this")
}

func Implement() (err error) {
	ui.TodoP1("start logging this")
	return errors.WrapSkip(1, errors.ErrNotImplemented)
}

func Remove() {
	ui.TodoP1("start logging this")
}
