package merkle

import (
	"encoding/hex"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

type IdStringerSetter Id

func (id IdStringerSetter) String() string {
	if id.tipe == "" && len(id.data) == 0 {
		return ""
	}

	if id.tipe == HRPObjectBlobDigestSha256V0 {
		if len(id.data) == 0 {
			return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		} else {
			return fmt.Sprintf("%x", id.data)
		}
	} else {
		bites, err := blech32.Encode(id.tipe, id.data)
		errors.PanicIfError(err)
		return string(bites)
	}
}

func (id *IdStringerSetter) SetMaybeSha256(value string) (err error) {
	if len(value) == 64 {
		if err = id.SetSha256(value); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = id.Set(value); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (id *IdStringerSetter) SetSha256(value string) (err error) {
	var decodedBytes []byte

	if decodedBytes, err = hex.DecodeString(value); err != nil {
		err = errors.Wrapf(err, "%q", value)
		return
	}

	if err = ((*Id)(id)).SetMerkleId(
		HRPObjectBlobDigestSha256V0,
		decodedBytes,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *IdStringerSetter) Set(value string) (err error) {
	if id.tipe, id.data, err = blech32.DecodeString(value); err != nil {
		err = errors.Wrapf(err, "Value: %q", value)
		return
	}

	if err = ((*Id)(id)).SetMerkleId(id.tipe, id.data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
