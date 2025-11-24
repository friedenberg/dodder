package markl

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
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

type KeyValueTuple[
	KEY interfaces.Stringer,
	KEY_PTR interfaces.StringerSetterPtr[KEY],
] struct {
	Key   KEY
	Value Id
}

func (tuple KeyValueTuple[KEY, KEY_PTR]) GetBinaryMarshaler() KeyValueTupleBinaryMarshaler[KEY, KEY_PTR] {
	return KeyValueTupleBinaryMarshaler[KEY, KEY_PTR](tuple)
}

type KeyValueTupleBinaryMarshaler[
	KEY interfaces.Stringer,
	KEY_PTR interfaces.StringerSetterPtr[KEY],
] KeyValueTuple[KEY, KEY_PTR]

func (tuple KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) MarshalBinary() (data []byte, err error) {
	return tuple.AppendBinary(nil)
}

func (tuple KeyValueTupleBinaryMarshaler[KEY, KEY_PTR]) AppendBinary(
	bites []byte,
) ([]byte, error) {
	bites = fmt.Append(bites, tuple.Key.String())

	if tuple.Value.IsEmpty() {
		return bites, errors.Errorf("empty type signature for %q", tuple.Key)
	}

	bites = append(bites, '\x00')
	bites = fmt.Append(bites, tuple.Value.GetMarklFormat().GetMarklFormatId())
	bites = append(bites, '\x00')
	bites = append(bites, tuple.Value.GetBytes()...)

	return bites, nil
}

type KeyValueTupleMutable[
	KEY interfaces.Stringer,
	KEY_PTR interfaces.StringerSetterPtr[KEY],
] struct {
	Key   KEY_PTR
	Value *Id
}

func (tuple KeyValueTupleMutable[KEY, KEY_PTR]) GetBinaryUnmarshaler() KeyValueTupleMutableBinaryUnmarshaler[KEY, KEY_PTR] {
	return KeyValueTupleMutableBinaryUnmarshaler[KEY, KEY_PTR](tuple)
}

type KeyValueTupleMutableBinaryUnmarshaler[
	KEY interfaces.Stringer,
	KEY_PTR interfaces.StringerSetterPtr[KEY],
] KeyValueTupleMutable[KEY, KEY_PTR]

func (tuple KeyValueTupleMutableBinaryUnmarshaler[KEY, KEY_PTR]) UnmarshalBinary(
	bites []byte,
) (err error) {
	if len(bites) == 0 {
		return
	}

	var formatAndBytes []byte

	{
		var key []byte
		var ok bool

		key, formatAndBytes, ok = bytes.Cut(bites, []byte{'\x00'})

		if !ok {
			if err = tuple.Key.Set(string(bites)); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

		if err = tuple.Key.Set(string(key)); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	{
		format, valueBytes, ok := bytes.Cut(formatAndBytes, []byte{'\x00'})

		if !ok {
			err = errors.Errorf("expected empty byte, but none found")
			return
		}

		id := tuple.Value
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

	return
}
