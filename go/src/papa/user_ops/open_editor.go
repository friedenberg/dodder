package user_ops

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/editor"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type OpenEditor struct {
	VimOptions []string
}

func (c OpenEditor) Run(
	u *local_working_copy.Repo,
	args ...string,
) (err error) {
	var e editor.Editor

	if e, err = editor.MakeEditorWithVimOptions(
		u.PrinterHeader(),
		c.VimOptions,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = e.Run(args); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
