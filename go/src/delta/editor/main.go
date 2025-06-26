package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/primordial"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"github.com/google/shlex"
)

type Editor struct {
	utility string
	path    string
	name    string
	tipe    Type

	options []string

	ui interfaces.FuncIter[string]
}

func getEditorUtility() string {
	var ed string

	if ed = os.Getenv("EDITOR"); ed != "" {
		return ed
	}

	if ed = os.Getenv("VISUAL"); ed != "" {
		return ed
	}

	return "vim"
}

func MakeEditorWithVimOptions(
	ph interfaces.FuncIter[string],
	options []string,
) (Editor, error) {
	return MakeEditor(
		ph,
		map[Type][]string{
			TypeVim: options,
		},
	)
}

func MakeEditor(
	ph interfaces.FuncIter[string],
	options map[Type][]string,
) (editor Editor, err error) {
	editor.utility = getEditorUtility()
	editor.ui = ph

	var utility []string

	if utility, err = shlex.Split(editor.utility); err != nil {
		err = errors.Wrap(err)
		return
	}

	if len(utility) < 1 {
		err = errors.ErrorWithStackf("utility has no valid path: %q", editor.utility)
		return
	}

	editor.path = utility[0]
	editor.options = append(editor.options, utility[1:]...)

	editor.name = filepath.Base(editor.path)

	switch editor.name {
	case "vim", "nvim":
		editor.tipe = TypeVim
		editor.options = append(editor.options, "-f")
	}

	editor.options = append(editor.options, options[editor.tipe]...)

	return
}

func (c Editor) Run(
	files []string,
) (err error) {
	if err = c.ui(fmt.Sprintf("editor (%s) started", c.name)); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = c.openWithArgs(files...); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = c.ui(fmt.Sprintf("editor (%s) closed", c.name)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (editor Editor) openWithArgs(fs ...string) (err error) {
	if len(fs) == 0 {
		err = errors.Wrap(files.ErrEmptyFileList)
		return
	}

	allArgs := append(editor.options, fs...)

	cmd := exec.Command(
		editor.path,
		allArgs...,
	)

	if primordial.IsTty(os.Stdin) {
		cmd.Stdin = os.Stdin
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		err = errors.Wrapf(err, "Cmd: %s", cmd)
		return
	}

	return
}
