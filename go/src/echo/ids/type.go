package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

func init() {
	register(Type{})
}

type (
	Type struct {
		Value string
	}

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

func (t Type) IsEmpty() bool {
	return t.Value == ""
}

func (t *Type) Reset() {
	t.Value = ""
}

func (a *Type) ResetWith(b Type) {
	a.Value = b.Value
}

func (a Type) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a Type) Equals(b Type) bool {
	return a.Value == b.Value
}

func (t Type) GetType() Type {
	return t
}

func (t *Type) GetTypPtr() *Type {
	return t
}

func (o Type) GetGenre() interfaces.Genre {
	return genres.Type
}

func (t Type) IsToml() bool {
	return strings.HasPrefix(t.Value, "toml")
}

func (e Type) StringSansOp() string {
	if e.IsEmpty() {
		return ""
	} else {
		return e.Value
	}
}

func (e Type) String() string {
	if e.IsEmpty() {
		return ""
	} else {
		return "!" + e.Value
	}
}

func (t Type) Parts() [3]string {
	return [3]string{"", "!", t.Value}
}

func (e *Type) TodoSetFromObjectId(v *ObjectId) (err error) {
	return e.Set(v.String())
}

func (e *Type) Set(v string) (err error) {
	v = strings.ToLower(strings.TrimSpace(strings.Trim(v, ".! ")))

	if err = ErrOnConfig(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !TagRegex.Match([]byte(v)) {
		err = errors.ErrorWithStackf("not a valid Typ: '%s'", v)
		return
	}

	e.Value = v

	return
}

func (t Type) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return
}

func (t *Type) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (t Type) MarshalBinary() (text []byte, err error) {
	text = []byte(t.String())
	return
}

func (t *Type) UnmarshalBinary(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
