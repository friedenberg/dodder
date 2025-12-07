package ids

import (
	"fmt"
	"strings"
	"unicode"

	"code.linenisgreat.com/dodder/go/src/_/coordinates"
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

func init() {
	register(ZettelId{})
}

type ZettelId struct {
	left, right string
}

type Provider interface {
	MakeZettelIdFromCoordinates(i coordinates.Int) (string, error)
}

func NewZettelIdEmpty() (h ZettelId) {
	h = ZettelId{}

	return h
}

// TODO-P3 is this really necessary?;w
func MakeZettelIdFromProvidersAndCoordinates(
	i coordinates.Int,
	pl Provider,
	pr Provider,
) (h *ZettelId, err error) {
	k := coordinates.ZettelIdCoordinate{}
	k.SetInt(i)

	var l, r string

	if l, err = pl.MakeZettelIdFromCoordinates(k.Left); err != nil {
		err = errors.ErrorWithStackf("failed to make left zettel id: %s", err)
		return h, err
	}

	if r, err = pr.MakeZettelIdFromCoordinates(k.Right); err != nil {
		err = errors.ErrorWithStackf("failed to make right zettel id: %s", err)
		return h, err
	}

	return MakeZettelIdFromHeadAndTail(l, r)
}

func MakeZettelIdFromHeadAndTail(head, tail string) (h *ZettelId, err error) {
	head = strings.TrimSpace(head)
	tail = strings.TrimSpace(tail)

	switch {
	case head == "":
		err = errors.ErrorWithStackf(
			"head was empty: {head: %q, tail: %q}",
			head,
			tail,
		)
		return h, err

	case tail == "":
		err = errors.ErrorWithStackf(
			"tail was empty: {head: %q, tail: %q}",
			head,
			tail,
		)
		return h, err
	}

	hs := fmt.Sprintf("%s/%s", head, tail)

	h = &ZettelId{}

	if err = h.Set(hs); err != nil {
		err = errors.ErrorWithStackf("failed to set zettel id: %s", err)
		return h, err
	}

	return h, err
}

func MustZettelId(v string) (h *ZettelId) {
	var err error
	h, err = MakeZettelId(v)

	errors.PanicIfError(err)

	return h
}

func MakeZettelId(v string) (h *ZettelId, err error) {
	h = &ZettelId{}

	if err = h.Set(v); err != nil {
		return h, err
	}

	return h, err
}

func (id ZettelId) IsEmpty() bool {
	return id.left == "" && id.right == ""
}

func (id ZettelId) EqualsAny(b any) bool {
	return values.Equals(id, b)
}

func (id ZettelId) Equals(b ZettelId) bool {
	if id.left != b.left {
		return false
	}

	if id.right != b.right {
		return false
	}

	return true
}

func (id ZettelId) GetHead() string {
	return id.left
}

func (id ZettelId) GetTail() string {
	return id.right
}

func (id ZettelId) GetObjectIdString() string {
	return id.String()
}

func (id ZettelId) String() string {
	v := fmt.Sprintf("%s/%s", id.left, id.right)
	return v
}

func (id ZettelId) Parts() [3]string {
	return [3]string{id.left, "/", id.right}
}

func (id ZettelId) Less(j ZettelId) bool {
	return id.String() < j.String()
}

func (h *ZettelId) SetFromIdParts(parts [3]string) (err error) {
	h.left = parts[0]
	h.right = parts[2]
	return err
}

func (id *ZettelId) Set(v string) (err error) {
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)

	v = strings.TrimSuffix(v, ".zettel")

	groupBuilder := errors.MakeGroupBuilder()

	if strings.ContainsFunc(
		v,
		func(r rune) bool {
			switch {
			case unicode.IsDigit(r),
				unicode.IsLetter(r),
				r == '_',
				r == '/',
				r == '%':
				return false

			default:
				return true
			}
		},
	) {
		groupBuilder.Add(
			errors.Errorf("contains invalid characters: %q", v),
		)
	}

	if v == "/" {
		if groupBuilder.Len() > 0 {
			err = groupBuilder.GetError()
		}

		return err
	}

	parts := strings.Split(v, "/")
	count := len(parts)

	switch count {
	default:
		groupBuilder.Add(errors.Errorf(
			"zettel id needs exactly 2 components, but got %d: %q",
			count,
			v,
		))

	case 2:
		id.left = parts[0]
		id.right = parts[1]
	}

	if (len(id.left) == 0 && len(id.right) > 0) ||
		(len(id.right) == 0 && len(id.left) > 0) {
		groupBuilder.Add(errors.Errorf("incomplete zettel id: %s", id))
	}

	err = groupBuilder.GetError()

	return err
}

func (id *ZettelId) Reset() {
	id.left = ""
	id.right = ""
}

func (id *ZettelId) ResetWith(h1 ZettelId) {
	id.left = h1.left
	id.right = h1.right
}

func (id ZettelId) GetGenre() interfaces.Genre {
	return genres.Zettel
}

func (id ZettelId) MarshalText() (text []byte, err error) {
	text = []byte(id.String())
	return text, err
}

func (id *ZettelId) UnmarshalText(text []byte) (err error) {
	if err = id.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id ZettelId) MarshalBinary() (text []byte, err error) {
	text = []byte(id.String())
	return text, err
}

func (id *ZettelId) UnmarshalBinary(text []byte) (err error) {
	if err = id.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id ZettelId) ToType() TypeStruct {
	panic(errors.Err405MethodNotAllowed)
}

func (id ZettelId) ToSeq() doddish.Seq {
	return doddish.Seq{
		doddish.Token{
			TokenType: doddish.TokenTypeIdentifier,
			Contents:  []byte(id.left),
		},
		doddish.Token{
			TokenType: doddish.TokenTypeOperator,
			Contents:  []byte{'/'},
		},
		doddish.Token{
			TokenType: doddish.TokenTypeIdentifier,
			Contents:  []byte(id.right),
		},
	}
}
