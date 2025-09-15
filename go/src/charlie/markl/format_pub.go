package markl

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type (
	FuncFormatPubVerify func(pubkey, message, sig interfaces.MarklId) error

	FormatPub struct {
		Id   string
		Size int

		Verify FuncFormatPubVerify
	}
)

var _ interfaces.MarklFormat = FormatPub{}

func (format FormatPub) GetMarklFormatId() string {
	return format.Id
}

func (format FormatPub) GetSize() int {
	return format.Size
}
