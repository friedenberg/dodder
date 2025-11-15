package ids

import (
	"regexp"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

func init() {
	register(Tag{})
}

const TagRegexString = `^%?[-a-z0-9_]+$`

var TagRegex *regexp.Regexp

func init() {
	TagRegex = regexp.MustCompile(TagRegexString)
}

var (
	sTagResetter  tagResetter
	sTag2Resetter tag2Resetter
)

type Tag = tag

var TagResetter = sTagResetter

// type Tag = tag2
// var TagResetter = sTag2Resetter

type tag struct {
	virtual       bool
	dependentLeaf bool
	value         string
}

func MustTagPtr(v string) (e *Tag) {
	e = &Tag{}
	e.init()

	var err error

	if err = e.Set(v); err != nil {
		errors.PanicIfError(err)
	}

	return e
}

func MustTag(v string) (e Tag) {
	e.init()

	var err error

	if err = e.Set(v); err != nil {
		errors.PanicIfError(err)
	}

	return e
}

func MakeTag(v string) (e Tag, err error) {
	e.init()

	if err = e.Set(v); err != nil {
		err = errors.Wrap(err)
		return e, err
	}

	return e, err
}

func (e tag) init() {
}

func (e *tag) Reset() {
	sTagResetter.Reset(e)
}

func (e tag) GetQueryPrefix() string {
	return "-"
}

func (e tag) IsEmpty() bool {
	return e.value == ""
}

func (e tag) GetGenre() interfaces.Genre {
	return genres.Tag
}

func (a tag) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a tag) Equals(b tag) bool {
	return a == b
}

func (i Tag) GetObjectIdString() string {
	return i.String()
}

func (e tag) String() string {
	var sb strings.Builder

	if e.virtual {
		sb.WriteRune('%')
	}

	if e.dependentLeaf {
		sb.WriteRune('-')
	}

	sb.WriteString(e.value)

	return sb.String()
}

func (e tag) Bytes() []byte {
	return []byte(e.String())
}

func (e tag) Parts() [3]string {
	switch {
	case e.virtual && e.dependentLeaf:
		return [3]string{"%", "-", e.value}

	case e.virtual:
		return [3]string{"", "%", e.value}

	case e.dependentLeaf:
		return [3]string{"", "-", e.value}

	default:
		return [3]string{"", "", e.value}
	}
}

func (tag tag) IsDodderTag() bool {
	return strings.HasPrefix(tag.value, "dodder-")
}

func (e tag) IsVirtual() bool {
	return e.virtual
}

func (e tag) IsDependentLeaf() bool {
	return e.dependentLeaf
}

func (e *tag) TodoSetFromObjectId(v *ObjectId) (err error) {
	return e.Set(v.String())
}

func (e *tag) Set(v string) (err error) {
	v1 := v
	v = strings.ToLower(strings.TrimSpace(v))

	if err = ErrOnConfig(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if !TagRegex.MatchString(v) {
		if v == "" {
			err = ErrEmptyTag
		} else {
			err = errors.ErrorWithStackf("not a valid tag: %q", v1)
		}

		return err
	}

	e.virtual = strings.HasPrefix(v, "%")
	v = strings.TrimPrefix(v, "%")

	e.dependentLeaf = strings.HasPrefix(v, "-")
	v = strings.TrimPrefix(v, "-")

	e.value = v

	return err
}

func (t tag) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *tag) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (t tag) MarshalBinary() (text []byte, err error) {
	text = []byte(t.String())
	return text, err
}

func (t *tag) UnmarshalBinary(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
