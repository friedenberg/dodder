package markl

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type KeyValueTuple[
	KEY interface {
		interfaces.Stringer
		interfaces.Equatable[KEY]
	},
	KEY_PTR interface {
		interfaces.Resetable
		interfaces.ResetablePtr[KEY]
		interfaces.StringerSetterPtr[KEY]
	},
] struct {
	Key   KEY
	Value Id
}

var _ interfaces.Resetable = &KeyValueTuple[ids.Type, *ids.Type]{}

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

func (tuple *KeyValueTuple[KEY, KEY_PTR]) GetBinaryMarshaler() KeyValueTupleBinaryMarshaler[KEY, KEY_PTR] {
	return KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]{tuple: tuple}
}

type KeyValueTupleBinaryMarshaler[
	KEY interface {
		interfaces.Stringer
		interfaces.Equatable[KEY]
	},
	KEY_PTR interface {
		interfaces.Resetable
		interfaces.ResetablePtr[KEY]
		interfaces.StringerSetterPtr[KEY]
	},
] struct {
	tuple *KeyValueTuple[KEY, KEY_PTR]
}

func (marshaler KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) MarshalBinary() (data []byte, err error) {
	return marshaler.AppendBinary(nil)
}

func (marshaler KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) AppendBinary(
	bites []byte,
) ([]byte, error) {
	bites = fmt.Append(bites, marshaler.tuple.Key.String())

	// if marshaler.tuple.Value.IsEmpty() {
	// 	return bites, errors.Errorf("empty type signature for %q", marshaler.tuple.Key)
	// }

	// bites = append(bites, '\x00')
	// bites = fmt.Append(bites, marshaler.tuple.Value.GetMarklFormat().GetMarklFormatId())
	// bites = append(bites, '\x00')
	// bites = append(bites, marshaler.tuple.Value.GetBytes()...)

	return bites, nil
}

func (marshaler KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) UnmarshalBinary(
	bites []byte,
) (err error) {
	if len(bites) == 0 {
		return err
	}

	// var formatAndBytes []byte

	{
		var key []byte
		var ok bool

		key, _, ok = bytes.Cut(bites, []byte{'\x00'})

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

	// {
	// 	format, valueBytes, ok := bytes.Cut(formatAndBytes, []byte{'\x00'})

	// 	if !ok {
	// 		err = errors.Errorf("expected empty byte, but none found")
	// 		return err
	// 	}

	// 	id := &marshaler.tuple.Value
	// 	id.Reset()

	// 	if err = id.setFormatId(string(format)); err != nil {
	// 		err = errors.Wrap(err)
	// 		return err
	// 	}

	// 	if err = id.setData(valueBytes); err != nil {
	// 		err = errors.Wrap(err)
	// 		return err
	// 	}
	// }

	return err
}
