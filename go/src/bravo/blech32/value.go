package blech32

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

type Value struct {
	HRP  string // human-readable part
	Data []byte
}

func MakeValue(
	hrp string,
	data []byte,
) Value {
	return Value{
		HRP:  hrp,
		Data: data,
	}
}

func MakeValueWithExpectedHRP(
	expectedHRP string,
	input string,
) (value Value, err error) {
	if err = value.Set(input); err != nil {
		err = errors.Wrap(err)
		return
	}

	if value.HRP != expectedHRP {
		err = errors.Errorf(
			"expected HRP %q but got %q",
			expectedHRP,
			value.HRP,
		)
		return
	}

	return
}

func (value Value) GetHRP() string {
	return value.HRP
}

func (value Value) GetData() []byte {
	return value.Data
}

func (value Value) String() string {
	var text []byte
	var err error

	if text, err = Encode(value.HRP, value.Data); err != nil {
		panic(err)
	}

	return string(text)
}

func (value *Value) Set(text string) (err error) {
	if len(text) == 0 {
		return
	}

	if value.HRP, value.Data, err = Decode(text); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (value Value) MarshalText() (text []byte, err error) {
	if len(value.Data) == 0 {
		return
	}

	if text, err = Encode(value.HRP, value.Data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (value *Value) UnmarshalText(text []byte) (err error) {
	if len(text) == 0 {
		return
	}

	if value.HRP, value.Data, err = Decode(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
