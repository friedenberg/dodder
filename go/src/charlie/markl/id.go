package markl

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

var (
	_ interfaces.MarklId        = Id{}
	_ interfaces.MutableMarklId = &Id{}
)

type Id struct {
	purpose string
	format  interfaces.MarklFormat
	data    []byte
}

func (id Id) String() string {
	if id.format == nil && len(id.data) == 0 {
		return ""
	}

	if len(id.data) == 0 {
		return ""
	} else {
		bites, err := blech32.Encode(id.format.GetMarklFormatId(), id.data)
		errors.PanicIfError(err)
		return string(bites)
	}
}

func (id Id) StringWithFormat() string {
	if id.format == nil && len(id.data) == 0 {
		return ""
	}

	if len(id.data) == 0 {
		return ""
	} else {
		bites, err := blech32.Encode(id.format.GetMarklFormatId(), id.data)
		bitesString := string(bites)
		errors.PanicIfError(err)

		if id.purpose != "" {
			return fmt.Sprintf("%s@%s", id.purpose, bitesString)
		} else {
			return bitesString
		}
	}
}

func (id Id) GetPurpose() string {
	return id.purpose
}

func (id *Id) SetPurpose(value string) error {
	id.purpose = value
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

func (id Id) GetMarklFormat() interfaces.MarklFormat {
	return id.format
}

func (id Id) IsNull() bool {
	if len(id.data) == 0 {
		return true
	} else if id.format == nil {
		panic("empty type")
	}

	formatHash, ok := formatHashes[id.format.GetMarklFormatId()]

	// this is not an Id for a hash, so it can never be null with non-zero data
	// contents
	if !ok {
		return false
	}

	if bytes.Equal(id.data, formatHash.null.GetBytes()) {
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
	if err = id.SetPurpose(format); err != nil {
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

	if err = id.SetMarklId(typeId, id.data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) SetDigest(digest interfaces.MarklId) (err error) {
	if err = id.SetPurpose(digest.GetPurpose()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetMarklId(
		digest.GetMarklFormat().GetMarklFormatId(),
		digest.GetBytes(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) setFormatId(formatId string) (err error) {
	if id.format, err = GetFormatOrError(formatId); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) SetMarklId(formatId string, bites []byte) (err error) {
	if formatId == "" && len(bites) == 0 {
		id.Reset()
		return
	}

	if err = id.setFormatId(formatId); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.setData(bites); err != nil {
		err = errors.Wrap(err)
		return
	}

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

func (id *Id) setData(bites []byte) (err error) {
	// empty is permitted
	if len(bites) == 0 {
		return
	}

	// TODO enforce non-nil formats
	if id.format != nil {
		expected := id.format.GetSize()
		actual := len(bites)

		if actual != expected {
			err = errors.Errorf(
				"wrong size for bytes: expected %d, but got %d",
				expected,
				actual,
			)

			return
		}
	}

	id.allocDataAndSetToCapIfNecessary(len(bites))
	copy(id.data, bites)

	return
}

func (id *Id) Reset() {
	id.format = nil
	id.data = id.data[:0]
	id.purpose = ""
}

func (id *Id) ResetWith(src Id) {
	id.purpose = src.purpose
	id.format = src.format
	errors.PanicIfError(id.setData(src.data))
}

func (id *Id) ResetWithMarklId(src interfaces.MarklId) {
	marklType := src.GetMarklFormat()
	bites := src.GetBytes()

	var marklTypeId string

	if marklType == nil && len(bites) > 0 {
		panic(
			fmt.Sprintf(
				"markl id with empty format but populated bytes: (bites: %x)",
				bites,
			),
		)
	} else if marklType != nil {
		marklTypeId = marklType.GetMarklFormatId()
	}

	errors.PanicIfError(
		id.SetPurpose(src.GetPurpose()),
	)

	errors.PanicIfError(
		id.SetMarklId(marklTypeId, bites),
	)
}

func (id *Id) GetBlobId() interfaces.MarklId {
	return id
}
