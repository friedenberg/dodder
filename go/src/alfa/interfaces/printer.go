package interfaces

import (
	"io"
	"os"
)

type Printer interface {
	io.Writer

	GetFile() *os.File
	IsTty() bool
	Print(a ...any) (err error)
	PrintDebug(a ...any) (err error)
	Printf(f string, a ...any) (err error)
}
