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

	formatIdBytes, bytesAfterFormatId, ok := bytes.Cut(bites, []byte{'\x00'})

	if !ok {
		err = errors.Errorf("expected empty byte, but none found")
		return
	}

	if err = id.SetMarklId(string(formatIdBytes), bytesAfterFormatId); err != nil {
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

	formatIdBytes, bytesAfterFormatId, ok := bytes.Cut(
		bytesAfterFormat,
		[]byte{'\x00'},
	)

	if !ok {
		err = errors.Errorf("expected empty byte, but none found")
		return
	}

	if err = id.SetMarklId(string(formatIdBytes), bytesAfterFormatId); err != nil {
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
	format := id.GetMarklFormat()
	bites := id.GetBytes()

	if format == nil && len(bites) == 0 {
		return
	} else if format == nil {
		err = errors.Errorf("empty type")
		return
	}

	bytes = append(bytes, []byte(format.GetMarklFormatId())...)
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
	purpose := id.GetPurpose()
	format := id.GetMarklFormat()
	bites := id.GetBytes()

	if format == nil && len(bites) == 0 && purpose == "" {
		return
	} else if format == nil {
		err = errors.Errorf("empty type")
		return
	}

	bytes = append(bytes, []byte(purpose)...)
	bytes = append(bytes, '\x00')
	bytes = append(bytes, []byte(format.GetMarklFormatId())...)
	bytes = append(bytes, '\x00')
	bytes = append(bytes, bites...)

	return
}
