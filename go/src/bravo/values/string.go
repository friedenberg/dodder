package values

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

type String struct {
	wasSet bool
	string
}

func MakeString(v string) String {
	return String{
		wasSet: true,
		string: v,
	}
}

func MakeStringDefault(v string) String {
	return String{
		string: v,
	}
}

func (str *String) Set(v string) (err error) {
	*str = String{
		wasSet: true,
		string: v,
	}

	return
}

func (str String) Match(v string) (err error) {
	if str.string != v {
		err = errors.BadRequestf("expected %q but got %q", str.string, v)
		return
	}

	return
}

func (str String) String() string {
	return str.string
}

func (str String) IsEmpty() bool {
	return len(str.string) == 0
}

func (str String) Len() int {
	return len(str.string)
}

func (str String) Less(other String) bool {
	return str.string < other.string
}

func (str String) WasSet() bool {
	return str.wasSet
}

func (str *String) Reset() {
	str.wasSet = false
	str.string = ""
}

func (a *String) ResetWith(b String) {
	a.wasSet = true
	a.string = b.string
}

func (s String) MarshalBinary() (text []byte, err error) {
	text = []byte(s.String())

	return
}

func (s *String) UnmarshalBinary(text []byte) (err error) {
	if err = s.Set(string(text)); err != nil {
		return
	}

	return
}
