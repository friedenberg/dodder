package markl

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func NonceGenerate(rand io.Reader) (bites []byte, err error) {
	bites = make([]byte, 32)

	if _, err = rand.Read(bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
