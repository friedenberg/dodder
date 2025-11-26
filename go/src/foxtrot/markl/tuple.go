package markl

import (
	"bytes"
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
)

// TODO rename to lock
type KeyValueTuple[
	KEY interfaces.Value[KEY],
	KEY_PTR interfaces.ValuePtr[KEY],
] struct {
	Key   KEY
	Value Id
}

var _ interfaces.Resetable = &KeyValueTuple[values.String, *values.String]{}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) GetKeyMutable() KEY_PTR {
	return KEY_PTR(&tuple.Key)
}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) Reset() {
	tuple.GetKeyMutable().Reset()
	tuple.Value.Reset()
}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) ResetWith(
	other KeyValueTuple[KEY, KEY_PTR],
) {
	tuple.GetKeyMutable().ResetWith(other.Key)
	tuple.Value.ResetWithMarklId(other.Value)
}

func (tuple KeyValueTuple[KEY, KEY_PTR]) IsEmpty() bool {
	return tuple.Key.IsEmpty() && tuple.Value.IsEmpty()
}

func (tuple KeyValueTuple[KEY, KEY_PTR]) Equals(
	other KeyValueTuple[KEY, KEY_PTR],
) bool {
	if !tuple.Key.Equals(other.Key) {
		return false
	}

	if !Equals(tuple.Value, other.Value) {
		return false
	}

	return true
}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) Set(
	value string,
) (err error) {
	key := tuple.GetKeyMutable()

	left, right, ok := strings.Cut(value, "@")

	if !ok {
		if err = key.Set(value); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}

	if err = key.Set(left); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = tuple.Value.Set(right); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) String() string {
	if tuple.Value.IsEmpty() {
		return tuple.Key.String()
	} else {
		return fmt.Sprintf("%s@%s", tuple.Key, tuple.Value)
	}
}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) GetBinaryMarshaler(
	requireValue bool,
) KeyValueTupleBinaryMarshaler[KEY, KEY_PTR] {
	return KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]{
		requireValue: requireValue,
		tuple:        tuple,
	}
}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) GetBinaryMarshalerValueNotRequired() KeyValueTupleBinaryMarshaler[KEY, KEY_PTR] {
	return tuple.GetBinaryMarshaler(false)
}

func (tuple *KeyValueTuple[KEY, KEY_PTR]) GetBinaryMarshalerValueRequired() KeyValueTupleBinaryMarshaler[KEY, KEY_PTR] {
	return tuple.GetBinaryMarshaler(true)
}

type KeyValueTupleBinaryMarshaler[
	KEY interfaces.Value[KEY],
	KEY_PTR interfaces.ValuePtr[KEY],
] struct {
	requireValue bool
	tuple        *KeyValueTuple[KEY, KEY_PTR]
}

func (marshaler KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) MarshalBinary() (data []byte, err error) {
	return marshaler.AppendBinary(nil)
}

func (marshaler KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) AppendBinary(
	bites []byte,
) ([]byte, error) {
	bites = fmt.Append(bites, marshaler.tuple.Key.String())

	if marshaler.tuple.Value.IsEmpty() {
		var err error

		if marshaler.requireValue {
			err = errors.Errorf("empty type signature for %q", marshaler.tuple.Key)
		}

		return bites, err
	}

	bites = append(bites, '\x00')
	formatId := marshaler.tuple.Value.GetMarklFormat().GetMarklFormatId()
	bites = fmt.Append(bites, formatId)
	bites = append(bites, '\x00')
	bites = append(bites, marshaler.tuple.Value.GetBytes()...)

	return bites, nil
}

func (marshaler KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) UnmarshalBinary(
	bites []byte,
) (err error) {
	if len(bites) == 0 {
		return err
	}

	var formatAndBytes []byte

	{
		var key []byte
		var ok bool

		key, formatAndBytes, ok = bytes.Cut(bites, []byte{'\x00'})

		if !ok {
			if err = marshaler.tuple.GetKeyMutable().Set(string(bites)); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		}

		if err = marshaler.tuple.GetKeyMutable().Set(string(key)); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	{
		format, valueBytes, ok := bytes.Cut(formatAndBytes, []byte{'\x00'})

		if !ok {
			err = errors.Errorf("expected empty byte between format id and id bytes")
			return err
		}

		id := &marshaler.tuple.Value
		id.Reset()

		if err = id.setFormatId(string(format)); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = id.setData(valueBytes); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
