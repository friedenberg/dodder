package interfaces

import (
	"io"
	"os"
)

type Printer interface {
	io.Writer

	GetFile() *os.File
	IsTty() bool
	// TODO add "isOn" function
	Print(a ...any) (err error)
	Caller(skip int) Printer
	PrintDebug(a ...any) (err error)
	Printf(f string, a ...any) (err error)
}
