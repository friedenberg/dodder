package typed_blob_store

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
)

type TomlBlobEncoder[
	O any,
	OPtr interfaces.Ptr[O],
] struct{}

func (TomlBlobEncoder[O, OPtr]) EncodeTo(
	t OPtr,
	w1 io.Writer,
) (n int64, err error) {
	w := bufio.NewWriter(w1)
	defer errors.DeferredFlusher(&err, w)

	enc := toml.NewEncoder(w)

	if err = enc.Encode(t); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}
