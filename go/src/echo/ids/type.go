package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

// TODO MAYBE move Type into its own package

func init() {
	register(Type{})
}

type (
	Type struct {
		Value string
	}

	// TODO rename to BinaryTypeChecker and flip uses
	InlineTypeChecker interface {
		IsInlineType(Type) bool
	}
)

func MakeType(v string) (t Type, err error) {
	if err = t.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MustType(v string) (t Type) {
	if err := t.Set(v); err != nil {
		errors.PanicIfError(err)
	}

	return
}

func (tipe Type) IsEmpty() bool {
	return tipe.Value == ""
}

func (tipe *Type) Reset() {
	tipe.Value = ""
}

func (tipe *Type) ResetWith(b Type) {
	tipe.Value = b.Value
}

func (tipe Type) EqualsAny(b any) bool {
	return values.Equals(tipe, b)
}

func (tipe Type) Equals(b Type) bool {
	return tipe.Value == b.Value
}

func (tipe Type) GetType() Type {
	return tipe
}

func (tipe *Type) GetTypPtr() *Type {
	return tipe
}

func (tipe Type) GetGenre() interfaces.Genre {
	return genres.Type
}

func (tipe Type) IsToml() bool {
	return strings.HasPrefix(tipe.Value, "toml")
}

func (tipe Type) StringSansOp() string {
	if tipe.IsEmpty() {
		return ""
	} else {
		return tipe.Value
	}
}

func (tipe Type) String() string {
	if tipe.IsEmpty() {
		return ""
	} else {
		return "!" + tipe.Value
	}
}

func (tipe Type) Parts() [3]string {
	return [3]string{"", "!", tipe.Value}
}

func (tipe *Type) TodoSetFromObjectId(v *ObjectId) (err error) {
	return tipe.Set(v.String())
}

func (tipe *Type) Set(v string) (err error) {
	v = strings.ToLower(strings.TrimSpace(strings.Trim(v, ".! ")))

	if err = ErrOnConfig(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !TagRegex.Match([]byte(v)) {
		err = errors.ErrorWithStackf("not a valid Typ: '%s'", v)
		return
	}

	tipe.Value = v

	return
}

func (tipe Type) MarshalText() (text []byte, err error) {
	text = []byte(tipe.String())
	return
}

func (tipe *Type) UnmarshalText(text []byte) (err error) {
	if err = tipe.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (tipe Type) MarshalBinary() (text []byte, err error) {
	text = []byte(tipe.String())
	return
}

func (tipe *Type) UnmarshalBinary(text []byte) (err error) {
	if err = tipe.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
