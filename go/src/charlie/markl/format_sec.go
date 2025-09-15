package markl

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type (
	// TODO switch to accepting bytes?
	FuncFormatSecGenerate     func(io.Reader) ([]byte, error)
	FuncFormatSecGetPublicKey func(private interfaces.MarklId) ([]byte, error)
	FuncFormatSecGetIOWrapper func(private interfaces.MarklId) (interfaces.IOWrapper, error)
	FuncFormatSecSign         func(sec, mes interfaces.MarklId, readerRand io.Reader) ([]byte, error)

	FormatSec struct {
		Id   string
		Size int

		Generate FuncFormatSecGenerate

		PubFormatId  string
		GetPublicKey FuncFormatSecGetPublicKey

		GetIOWrapper FuncFormatSecGetIOWrapper

		SigFormatId string
		Sign        FuncFormatSecSign
	}
)

var _ interfaces.MarklFormat = FormatSec{}

func (format FormatSec) GetMarklFormatId() string {
	return format.Id
}

func (format FormatSec) GetSize() int {
	return format.Size
}
