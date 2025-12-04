package ids

import (
	"regexp"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

func init() {
	register(TagStruct{})
}

const TagRegexString = `^%?[-a-z0-9_]+$`

var TagRegex *regexp.Regexp

func init() {
	TagRegex = regexp.MustCompile(TagRegexString)
}

var sTagResetter tagResetter

type (
	TagStruct = tag
	ITag      = interface {
		interfaces.ObjectId
	}
)

var (
	TagResetter           = sTagResetter
	_           TagStruct = TagStruct{}
)

type tag struct {
	virtual       bool
	dependentLeaf bool
	value         string
}

func MustTagPtr(value string) *TagStruct {
	tag := &TagStruct{}
	tag.init()

	var err error

	if err = tag.Set(value); err != nil {
		errors.PanicIfError(err)
	}

	return tag
}

func MustTag(value string) TagStruct {
	var tag TagStruct
	tag.init()

	var err error

	if err = tag.Set(value); err != nil {
		errors.PanicIfError(err)
	}

	return tag
}

func MakeTag(value string) (TagStruct, error) {
	var tag TagStruct
	tag.init()

	if err := tag.Set(value); err != nil {
		err = errors.Wrap(err)
		return tag, err
	}

	return tag, nil
}

func (tag tag) init() {
}

func (tag *tag) Reset() {
	sTagResetter.Reset(tag)
}

func (tag *tag) ResetWith(other tag) {
	sTagResetter.ResetWith(tag, &other)
}

func (tag tag) GetQueryPrefix() string {
	return "-"
}

func (tag tag) IsEmpty() bool {
	return tag.value == ""
}

func (tag tag) GetGenre() interfaces.Genre {
	return genres.Tag
}

func (tag tag) EqualsAny(b any) bool {
	return values.Equals(tag, b)
}

func (tag tag) Equals(b tag) bool {
	return tag == b
}

func (tag TagStruct) GetObjectIdString() string {
	return tag.String()
}

func (tag tag) String() string {
	var sb strings.Builder

	if tag.virtual {
		sb.WriteRune('%')
	}

	if tag.dependentLeaf {
		sb.WriteRune('-')
	}

	sb.WriteString(tag.value)

	return sb.String()
}

func (tag tag) Bytes() []byte {
	return []byte(tag.String())
}

func (tag tag) Parts() [3]string {
	switch {
	case tag.virtual && tag.dependentLeaf:
		return [3]string{"%", "-", tag.value}

	case tag.virtual:
		return [3]string{"", "%", tag.value}

	case tag.dependentLeaf:
		return [3]string{"", "-", tag.value}

	default:
		return [3]string{"", "", tag.value}
	}
}

func (tag tag) IsDodderTag() bool {
	return strings.HasPrefix(tag.value, "dodder-")
}

func TagIsVirtual(tag ITag) bool {
	// TODO panic if tag is not tag
	return strings.HasPrefix(tag.String(), "%")
}

func (tag tag) IsVirtual() bool {
	return tag.virtual
}

func (tag tag) IsDependentLeaf() bool {
	return tag.dependentLeaf
}

func (tag *tag) TodoSetFromObjectId(v *ObjectId) (err error) {
	return tag.Set(v.String())
}

func (tag *tag) Set(v string) (err error) {
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

	tag.virtual = strings.HasPrefix(v, "%")
	v = strings.TrimPrefix(v, "%")

	tag.dependentLeaf = strings.HasPrefix(v, "-")
	v = strings.TrimPrefix(v, "-")

	tag.value = v

	return err
}

func (tag tag) MarshalText() (text []byte, err error) {
	text = []byte(tag.String())
	return text, err
}

func (tag *tag) UnmarshalText(text []byte) (err error) {
	if err = tag.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (tag tag) MarshalBinary() (text []byte, err error) {
	text = []byte(tag.String())
	return text, err
}

func (tag *tag) UnmarshalBinary(text []byte) (err error) {
	if err = tag.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func IsDependentLeaf(a TagStruct) (has bool) {
	has = strings.HasPrefix(strings.TrimSpace(a.String()), "-")
	return has
}

func HasParentPrefix(a, b TagStruct) (has bool) {
	has = strings.HasPrefix(strings.TrimSpace(a.String()), b.String())
	return has
}

type tagResetter struct{}

func (tagResetter) Reset(e *TagStruct) {
	e.value = ""
	e.virtual = false
	e.dependentLeaf = false
}

func (tagResetter) ResetWith(a, b *TagStruct) {
	a.value = b.value
	a.virtual = b.virtual
	a.dependentLeaf = b.dependentLeaf
}
