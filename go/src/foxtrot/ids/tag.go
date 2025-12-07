package ids

import (
	"regexp"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
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

var (
	TagResetter     = sTagResetter
	_           Tag = TagStruct{}
)

type tagStruct struct {
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

func (tag tagStruct) init() {
}

func (tag *tagStruct) Reset() {
	sTagResetter.Reset(tag)
}

func (tag *tagStruct) ResetWith(other tagStruct) {
	sTagResetter.ResetWith(tag, &other)
}

func (tag tagStruct) GetQueryPrefix() string {
	return "-"
}

func (tag tagStruct) IsEmpty() bool {
	return tag.value == ""
}

func (tag tagStruct) GetGenre() interfaces.Genre {
	return genres.Tag
}

func (tag tagStruct) Equals(b tagStruct) bool {
	return tag == b
}

func (tag TagStruct) GetObjectIdString() string {
	return tag.String()
}

func (tag tagStruct) String() string {
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

func (tag tagStruct) Bytes() []byte {
	return []byte(tag.String())
}

func (tag tagStruct) Parts() [3]string {
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

func (tag tagStruct) IsDodderTag() bool {
	return strings.HasPrefix(tag.value, "dodder-")
}

func TagIsVirtual(tag Tag) bool {
	// TODO panic if tag is not tag
	return strings.HasPrefix(tag.String(), "%")
}

func (tag tagStruct) IsVirtual() bool {
	return tag.virtual
}

func (tag tagStruct) IsDependentLeaf() bool {
	return tag.dependentLeaf
}

func (tag *tagStruct) TodoSetFromObjectId(v *ObjectId) (err error) {
	return tag.Set(v.String())
}

func (tag *tagStruct) Set(v string) (err error) {
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

func (tag tagStruct) MarshalText() (text []byte, err error) {
	text = []byte(tag.String())
	return text, err
}

func (tag *tagStruct) UnmarshalText(text []byte) (err error) {
	if err = tag.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (tag tagStruct) MarshalBinary() (text []byte, err error) {
	text = []byte(tag.String())
	return text, err
}

func (tag *tagStruct) UnmarshalBinary(text []byte) (err error) {
	if err = tag.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (tag tagStruct) ToType() TypeStruct {
	panic(errors.Err405MethodNotAllowed)
}

func (tag tagStruct) ToSeq() doddish.Seq {
	switch {
	case tag.virtual && tag.dependentLeaf:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{'%'},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{'-'},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  []byte(tag.value),
			},
		}

	case tag.virtual:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{'%'},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  []byte(tag.value),
			},
		}

	case tag.dependentLeaf:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{'-'},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  []byte(tag.value),
			},
		}

	default:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  []byte(tag.value),
			},
		}
	}
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

func (tagResetter) Reset(tag *TagStruct) {
	tag.value = ""
	tag.virtual = false
	tag.dependentLeaf = false
}

func (tagResetter) ResetWith(dst, src *TagStruct) {
	dst.value = src.value
	dst.virtual = src.virtual
	dst.dependentLeaf = src.dependentLeaf
}
