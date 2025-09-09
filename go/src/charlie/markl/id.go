package markl

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var idPool interfaces.Pool[Id, *Id] = pool.MakeWithResetable[Id]()

func PutBlobId(digest interfaces.MarklId) {
	switch id := digest.(type) {
	case Id:
		idPool.Put(&id)

	case *Id:
		idPool.Put(id)

	default:
		panic(errors.Errorf("unsupported id type: %T", digest))
	}
}

var (
	_ interfaces.MarklId        = Id{}
	_ interfaces.MutableMarklId = &Id{}
)

type Id struct {
	format string
	tipe   interfaces.MarklType
	data   []byte
}

func (id Id) String() string {
	if id.tipe == nil && len(id.data) == 0 {
		return ""
	}

	if id.tipe.GetMarklTypeId() == HashTypeIdSha256 {
		return fmt.Sprintf("%x", id.data)
	} else if len(id.data) == 0 {
		return ""
	} else {
		bites, err := blech32.Encode(id.tipe.GetMarklTypeId(), id.data)
		errors.PanicIfError(err)
		return string(bites)
	}
}

func (id Id) StringWithFormat() string {
	if id.tipe == nil && len(id.data) == 0 {
		return ""
	}

	if len(id.data) == 0 {
		return ""
	} else {
		bites, err := blech32.Encode(id.tipe.GetMarklTypeId(), id.data)
		bitesString := string(bites)
		errors.PanicIfError(err)

		if id.format != "" {
			return fmt.Sprintf("%s@%s", id.format, bitesString)
		} else {
			return bitesString
		}
	}
}

func (id Id) GetFormat() string {
	return id.format
}

func (id *Id) SetFormat(value string) error {
	id.format = value
	return nil
}

func (id Id) IsEmpty() bool {
	return len(id.data) == 0
}

func (id Id) GetSize() int {
	return len(id.data)
}

func (id Id) GetBytes() []byte {
	return id.data
}

func (id Id) GetMarklType() interfaces.MarklType {
	return id.tipe
}

func (id Id) IsNull() bool {
	if len(id.data) == 0 {
		return true
	} else if id.tipe == nil {
		panic("empty type")
	}

	hashType, ok := hashTypes[id.tipe.GetMarklTypeId()]

	// this is not an Id for a hash, so it can never be null with non-zero data
	// contents
	if !ok {
		return false
	}

	if bytes.Equal(id.data, hashType.null.GetBytes()) {
		return true
	}

	return false
}

func (id *Id) Set(value string) (err error) {
	format, body, ok := strings.Cut(value, "@")

	if ok {
		if err = id.setWithFormat(format, body); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = id.setWithoutFormat(value); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (id *Id) setWithFormat(format, body string) (err error) {
	if err = id.SetFormat(format); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.Set(body); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) setWithoutFormat(value string) (err error) {
	var typeId string

	if typeId, id.data, err = blech32.DecodeString(value); err != nil {
		err = errors.Wrapf(err, "Value: %q", value)
		return
	}

	if err = id.SetMerkleId(typeId, id.data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) SetDigest(digest interfaces.MarklId) (err error) {
	if err = id.SetFormat(digest.GetFormat()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetMerkleId(
		digest.GetMarklType().GetMarklTypeId(),
		digest.GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) SetMerkleId(tipe string, bites []byte) (err error) {
	if tipe == "" && len(bites) == 0 {
		id.Reset()
		return
	}

	if id.tipe, err = GetMarklTypeOrError(tipe); err != nil {
		err = errors.Wrapf(
			err,
			"failed to SetMerkleId on %T with contents: %s",
			id,
			bites,
		)
		return
	}

	id.setData(bites)

	return
}

func (id *Id) allocDataIfNecessary(size int) {
	id.data = id.data[:0]
	id.data = slices.Grow(id.data, size)
}

func (id *Id) allocDataAndSetToCapIfNecessary(size int) {
	id.allocDataIfNecessary(size)
	id.data = id.data[:size]
}

func (id *Id) setData(data []byte) {
	// TODO validate against type
	id.allocDataAndSetToCapIfNecessary(len(data))
	copy(id.data, data)
}

func (id *Id) Reset() {
	id.tipe = nil
	id.data = id.data[:0]
	id.format = ""
}

func (id *Id) ResetWith(src Id) {
	id.format = src.format
	id.tipe = src.tipe
	id.setData(src.data)
}

func (id *Id) ResetWithMarklId(src interfaces.MarklId) {
	marklType := src.GetMarklType()
	bites := src.GetBytes()

	var marklTypeId string

	if marklType == nil && len(bites) > 0 {
		panic(
			fmt.Sprintf(
				"markl id with empty tipe but populated bytes: (bites: %x)",
				bites,
			),
		)
	} else if marklType != nil {
		marklTypeId = marklType.GetMarklTypeId()
	}

	errors.PanicIfError(
		id.SetFormat(src.GetFormat()),
	)

	errors.PanicIfError(
		id.SetMerkleId(marklTypeId, bites),
	)
}

func (id *Id) GetBlobId() interfaces.MarklId {
	return id
}
