package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

// TODO MAYBE move Type into its own package

func init() {
	register(typeStruct{})
}

type typeStruct struct {
	Value string
}

var _ Type = TypeStruct{}

func MakeTypeString(value string) string {
	value = strings.TrimSpace(value)

	if !strings.HasPrefix(value, "!") {
		value = "!" + value
	}

	return value
}

func MakeType(value string) (tipe SeqId, err error) {
	if err = tipe.SetType(value); err != nil {
		err = errors.Wrap(err)
		return tipe, err
	}

	if err = genres.Type.AssertGenre(tipe); err != nil {
		err = errors.Wrap(err)
		return tipe, err
	}

	return tipe, err
}

func MustType(value string) (tipe SeqId) {
	var err error
	tipe, err = MakeType(value)
	if err != nil {
		errors.PanicIfError(err)
	}

	return tipe
}

func MakeTypeStruct(value string) (tipe typeStruct, err error) {
	if err = tipe.Set(value); err != nil {
		err = errors.Wrap(err)
		return tipe, err
	}

	return tipe, err
}

func MustTypeStruct(value string) (tipe typeStruct) {
	if err := tipe.Set(value); err != nil {
		errors.PanicIfError(err)
	}

	return tipe
}

func (typeStruct typeStruct) IsEmpty() bool {
	return typeStruct.Value == ""
}

func (typeStruct *typeStruct) Reset() {
	typeStruct.Value = ""
}

func (typeStruct *typeStruct) ResetWith(b typeStruct) {
	typeStruct.Value = b.Value
}

func (typeStruct typeStruct) Equals(b typeStruct) bool {
	return typeStruct.Value == b.Value
}

func (typeStruct typeStruct) GetGenre() interfaces.Genre {
	return genres.Type
}

func (typeStruct typeStruct) StringSansOp() string {
	if typeStruct.IsEmpty() {
		return ""
	} else {
		return typeStruct.Value
	}
}

func (typeStruct typeStruct) String() string {
	if typeStruct.IsEmpty() {
		return ""
	} else {
		return "!" + typeStruct.Value
	}
}

func (typeStruct *typeStruct) TodoSetFromObjectId(v *ObjectId) (err error) {
	return typeStruct.Set(v.String())
}

func (typeStruct *typeStruct) Set(value string) (err error) {
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

func (typeStruct typeStruct) MarshalText() (text []byte, err error) {
	text = []byte(typeStruct.String())
	return text, err
}

func (typeStruct *typeStruct) UnmarshalText(text []byte) (err error) {
	if err = typeStruct.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (typeStruct typeStruct) MarshalBinary() (text []byte, err error) {
	text = []byte(typeStruct.String())
	return text, err
}

func (typeStruct *typeStruct) UnmarshalBinary(text []byte) (err error) {
	if err = typeStruct.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (typeStruct typeStruct) ToType() TypeStruct {
	return typeStruct
}

func (typeStruct typeStruct) ToSeq() doddish.Seq {
	return doddish.Seq{
		doddish.Token{
			TokenType: doddish.TokenTypeOperator,
			Contents:  []byte("!"),
		},
		doddish.Token{
			TokenType: doddish.TokenTypeIdentifier,
			Contents:  []byte(typeStruct.StringSansOp()),
		},
	}
}
