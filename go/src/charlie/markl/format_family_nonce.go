package markl

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func NonceGenerate(rand io.Reader, size int) (bites []byte, err error) {
	bites = make([]byte, size)

	if _, err = rand.Read(bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func NonceGenerate32(rand io.Reader) (bites []byte, err error) {
	return NonceGenerate(rand, 32)
}
