package merkle

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var (
	_    interfaces.BlobId = null{}
	Null                   = null{}
)

type null struct {
	tipe string
	data []byte
}

func (id null) String() string {
	return "null"
}

func (id null) IsEmpty() bool {
	return true
}

func (id null) GetSize() int {
	return 0
}

func (id null) GetBytes() []byte {
	return nil
}

func (id null) GetType() string {
	return ""
}

func (id null) IsNull() bool {
	return true
}

func (id null) MarshalBinary() (bytes []byte, err error) {
	err = errors.Err405MethodNotAllowed
	return
}

func (id null) MarshalText() (bites []byte, err error) {
	err = errors.Err405MethodNotAllowed
	return
}
