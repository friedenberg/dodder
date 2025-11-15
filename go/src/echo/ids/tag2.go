package ids

import (
	"regexp"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

const Tag2RegexString = `^%?[-a-z0-9_]+$`

var Tag2Regex *regexp.Regexp

func init() {
	Tag2Regex = regexp.MustCompile(Tag2RegexString)
}

type tag2 struct {
	virtual       bool
	dependentLeaf bool
	value         *catgut.String
}

func (e *tag2) init() {
	e.value = catgut.GetPool().Get()
}

func (e tag2) GetQueryPrefix() string {
	return "-"
}

func (e tag2) GetGenre() interfaces.Genre {
	return genres.Tag
}

func (e *tag2) Reset() {
	sTag2Resetter.Reset(e)
}

func (a tag2) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a tag2) Equals(b tag2) bool {
	return a == b
}

func (e tag2) String() string {
	var sb strings.Builder

	if e.virtual {
		sb.WriteRune('%')
	}

	if e.dependentLeaf {
		sb.WriteRune('-')
	}

	sb.Write(e.value.Bytes())

	return sb.String()
}

func (e tag2) Bytes() []byte {
	return []byte(e.String())
}

func (e tag2) Parts() [3]string {
	switch {
	case e.virtual && e.dependentLeaf:
		return [3]string{"%", "-", e.value.String()}

	case e.virtual:
		return [3]string{"", "%", e.value.String()}

	case e.dependentLeaf:
		return [3]string{"", "-", e.value.String()}

	default:
		return [3]string{"", "", e.value.String()}
	}
}

func (e tag2) IsVirtual() bool {
	return e.virtual
}

func (e tag2) IsDependentLeaf() bool {
	return e.dependentLeaf
}

func (e *tag2) TodoSetFromObjectId(v *ObjectId) (err error) {
	return e.Set(v.String())
}

func (e *tag2) Set(v string) (err error) {
	v1 := v
	v = strings.ToLower(strings.TrimSpace(v))

	if err = ErrOnConfig(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if !Tag2Regex.MatchString(v) {
		err = errors.ErrorWithStackf("not a valid tag: %q", v1)
		return err
	}

	e.virtual = strings.HasPrefix(v, "%")
	v = strings.TrimPrefix(v, "%")

	e.dependentLeaf = strings.HasPrefix(v, "-")
	v = strings.TrimPrefix(v, "-")

	if e.value == nil {
		e.value = catgut.GetPool().Get()
	}

	if err = e.value.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (t tag2) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *tag2) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (t tag2) MarshalBinary() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *tag2) UnmarshalBinary(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
