package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

// TODO MAYBE move Type into its own package

func init() {
	register(TypeStruct{})
}

type (
	TypeStruct struct {
		Value string
	}

	IType = TypeStruct
	// IType        = interfaces.ObjectId
	ITypeMutable = *TypeStruct

	// TODO rename to BinaryTypeChecker and flip uses
	InlineTypeChecker interface {
		IsInlineType(IType) bool
	}
)

func MakeType(value string) (tipe TypeStruct, err error) {
	if err = tipe.Set(value); err != nil {
		err = errors.Wrap(err)
		return tipe, err
	}

	return tipe, err
}

func MustType(value string) (tipe TypeStruct) {
	if err := tipe.Set(value); err != nil {
		errors.PanicIfError(err)
	}

	return tipe
}

func (typeStruct TypeStruct) IsEmpty() bool {
	return typeStruct.Value == ""
}

func (typeStruct *TypeStruct) Reset() {
	typeStruct.Value = ""
}

func (typeStruct *TypeStruct) ResetWith(b TypeStruct) {
	typeStruct.Value = b.Value
}

func (typeStruct TypeStruct) EqualsAny(b any) bool {
	return values.Equals(typeStruct, b)
}

func (typeStruct TypeStruct) Equals(b TypeStruct) bool {
	return typeStruct.Value == b.Value
}

func (typeStruct TypeStruct) GetType() TypeStruct {
	return typeStruct
}

func (typeStruct *TypeStruct) GetTypPtr() *TypeStruct {
	return typeStruct
}

func (typeStruct TypeStruct) GetGenre() interfaces.Genre {
	return genres.Type
}

func (typeStruct TypeStruct) IsToml() bool {
	return strings.HasPrefix(typeStruct.Value, "toml")
}

func (typeStruct TypeStruct) StringSansOp() string {
	if typeStruct.IsEmpty() {
		return ""
	} else {
		return typeStruct.Value
	}
}

func (typeStruct TypeStruct) String() string {
	if typeStruct.IsEmpty() {
		return ""
	} else {
		return "!" + typeStruct.Value
	}
}

func (typeStruct TypeStruct) Parts() [3]string {
	return [3]string{"", "!", typeStruct.Value}
}

func (typeStruct *TypeStruct) TodoSetFromObjectId(v *ObjectId) (err error) {
	return typeStruct.Set(v.String())
}

func (typeStruct *TypeStruct) Set(value string) (err error) {
	value = strings.ToLower(strings.TrimSpace(strings.Trim(value, ".! ")))

	if err = ErrOnConfig(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if !TagRegex.Match([]byte(value)) {
		err = errors.ErrorWithStackf("not a valid Type: '%s'", value)
		return err
	}

	typeStruct.Value = value

	return err
}

func (typeStruct TypeStruct) MarshalText() (text []byte, err error) {
	text = []byte(typeStruct.String())
	return text, err
}

func (typeStruct *TypeStruct) UnmarshalText(text []byte) (err error) {
	if err = typeStruct.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (typeStruct TypeStruct) MarshalBinary() (text []byte, err error) {
	text = []byte(typeStruct.String())
	return text, err
}

func (typeStruct *TypeStruct) UnmarshalBinary(text []byte) (err error) {
	if err = typeStruct.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
