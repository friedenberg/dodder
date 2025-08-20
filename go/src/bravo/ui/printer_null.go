package ui

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

var Null null

func init() {
	var err error
	Null.file, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	errors.PanicIfError(err)
}

type null struct {
	file *os.File
}

func (null) Write(b []byte) (n int, err error) {
	n = len(b)
	return
}

func (printer null) GetFile() *os.File {
	return printer.file
}

func (null) IsTty() bool {
	return false
}

func (null) PrintDebug(_ ...any) (err error) {
	return
}

func (null) Print(_ ...any) (err error) {
	return
}

func (null) Printf(_ string, _ ...any) (err error) {
	return
}
