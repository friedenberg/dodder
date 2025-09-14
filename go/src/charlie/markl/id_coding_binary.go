package markl

import (
	"bytes"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type IdBinaryDecodingTypeData struct {
	interfaces.MutableMarklId
}

func (id IdBinaryDecodingTypeData) UnmarshalBinary(
	bites []byte,
) (err error) {
	if len(bites) == 0 {
		return
	}

	tipeBytes, bytesAfterTipe, ok := bytes.Cut(bites, []byte{'\x00'})

	if !ok {
		err = errors.Errorf("expected empty byte, but none found")
		return
	}

	if err = id.SetMarklId(string(tipeBytes), bytesAfterTipe); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type IdBinaryDecodingFormatTypeData struct {
	interfaces.MutableMarklId
}

func (id *Id) UnmarshalBinary(
	bites []byte,
) (err error) {
	if len(bites) == 0 {
		return
	}

	formatBytes, bytesAfterFormat, ok := bytes.Cut(bites, []byte{'\x00'})

	if !ok {
		err = errors.Errorf("expected empty byte, but none found")
		return
	}

	if err = id.SetPurpose(string(formatBytes)); err != nil {
		err = errors.Wrap(err)
		return
	}

	tipeBytes, bytesAfterTipe, ok := bytes.Cut(bytesAfterFormat, []byte{'\x00'})

	if !ok {
		err = errors.Errorf("expected empty byte, but none found")
		return
	}

	if err = id.SetMarklId(string(tipeBytes), bytesAfterTipe); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type IdBinaryEncodingTypeData struct {
	interfaces.MarklId
}

// structure (in bytes):
// <256: type
// 1: empty byte
// <256: id
func (id IdBinaryEncodingTypeData) MarshalBinary() (bytes []byte, err error) {
	// TODO confirm few allocations
	// TODO confirm size of type is less than 256
	tipe := id.GetMarklFormat()
	bites := id.GetBytes()

	if tipe == nil && len(bites) == 0 {
		return
	} else if tipe == nil {
		err = errors.Errorf("empty type")
		return
	}

	bytes = append(bytes, []byte(tipe.GetMarklFormatId())...)
	bytes = append(bytes, '\x00')
	bytes = append(bytes, bites...)

	return
}

type IdBinaryEncodingFormatTypeData struct {
	interfaces.MarklId
}

// structure (in bytes):
// <256: type
// 1: empty byte
// <256: id
func (id Id) MarshalBinary() (bytes []byte, err error) {
	// TODO confirm few allocations
	// TODO confirm size of type is less than 256
	format := id.GetPurpose()
	tipe := id.GetMarklFormat()
	bites := id.GetBytes()

	if tipe == nil && len(bites) == 0 && format == "" {
		return
	} else if tipe == nil {
		err = errors.Errorf("empty type")
		return
	}

	bytes = append(bytes, []byte(format)...)
	bytes = append(bytes, '\x00')
	bytes = append(bytes, []byte(tipe.GetMarklFormatId())...)
	bytes = append(bytes, '\x00')
	bytes = append(bytes, bites...)

	return
}
