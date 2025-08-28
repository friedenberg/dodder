package merkle

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

var (
	_ interfaces.BlobId        = IdSupportingEmpty{}
	_ interfaces.MutableBlobId = &IdSupportingEmpty{}
)

type IdSupportingEmpty struct {
	Id
}

func (id *IdSupportingEmpty) Set(value string) (err error) {
	if id.tipe, id.data, err = blech32.DecodeString(value); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetMerkleId(id.tipe, id.data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *IdSupportingEmpty) UnmarshalText(bites []byte) (err error) {
	if id.tipe, id.data, err = blech32.Decode(bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
