package ids

import (
	"encoding"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

type (
	TagStruct  = tagStruct
	TypeStruct = typeStruct

	Id interface {
		interfaces.ObjectId
		encoding.BinaryMarshaler
		encoding.BinaryAppender

		ToSeq() doddish.Seq
		ToType() TypeStruct
		IsEmpty() bool
	}

	IdMutable interface {
		Id
		encoding.BinaryUnmarshaler
		// SetWithGenre(string, interfaces.GenreGetter) error
		SetWithSeq(doddish.Seq) error
	}

	Tag         = Id
	SeqId       = objectId3
	Type        = Id
	TypeMutable = IdMutable

	// TODO rename to BinaryTypeChecker and flip uses
	InlineTypeChecker interface {
		IsInlineType(Type) bool
	}

	ObjectIdGetter interface {
		GetObjectId() *ObjectId
	}
)

type ObjectId = objectId2

func GetObjectIdPool() interfaces.Pool[ObjectId, *ObjectId] {
	return getObjectIdPool2()
}

// type ObjectId = objectId3

// func GetObjectIdPool() interfaces.Pool[ObjectId, *ObjectId] {
// 	return getObjectIdPool3()
// }

var (
	_ Id        = &ObjectId{}
	_ IdMutable = &ObjectId{}
)

func MustObjectId(idWithParts Id) (id *ObjectId) {
	id = &ObjectId{}
	err := id.SetWithId(idWithParts)
	errors.PanicIfError(err)
	return id
}

func MakeObjectId(value string) (objectId *ObjectId, err error) {
	if value == "" {
		objectId = GetObjectIdPool().Get()
		return objectId, err
	}

	var seq doddish.Seq

	if seq, err = doddish.ScanExactlyOneSeqWithDotAllowedInIdenfierFromString(
		value,
	); err != nil {
		err = errors.Wrap(err)
		return objectId, err
	}

	objectId = GetObjectIdPool().Get()

	if err = objectId.SetWithSeq(seq); err != nil {
		return objectId, err
	}

	return objectId, err
}

// TODO rewrite to use ToSeq comparison
func Equals(left, right interfaces.ObjectId) (ok bool) {
	if left.GetGenre().String() != right.GetGenre().String() {
		return ok
	}

	if left.String() != right.String() {
		return ok
	}

	return true
}

func FormattedString(id Id) string {
	return id.ToSeq().String()
}

func LeftSubtract[
	ID interfaces.Stringer,
	ID_PTR interfaces.StringSetterPtr[ID],
](
	base, toSubtract ID,
) (c ID, err error) {
	if err = ID_PTR(&c).Set(strings.TrimPrefix(base.String(), toSubtract.String())); err != nil {
		err = errors.Wrapf(err, "'%s' - '%s'", base, toSubtract)
		return c, err
	}

	return c, err
}

func Contains(left, right Id) bool {
	var (
		leftSeq  = left.ToSeq()
		rightSeq = right.ToSeq()
	)

	return doddish.SeqComparePartial(leftSeq, rightSeq).IsEqual()
}

func ContainsExactly(left, right Id) bool {
	var (
		leftSeq  = left.ToSeq()
		rightSeq = right.ToSeq()
	)

	return doddish.SeqCompare(leftSeq, rightSeq).IsEqual()
}

func IsEmpty(id Id) bool {
	return id.ToSeq().Len() == 0
}

func IsVirtual(id Id) bool {
	seq := id.ToSeq()

	if seq.Len() < 1 {
		return false
	} else if seq.MatchStart(
		doddish.TokenMatcherOp(doddish.OpVirtual),
	) {
		return true
	} else {
		return false
	}
}

func Expand(
	id *ObjectId,
	abbr Abbr,
) (err error) {
	genre := genres.Must(id)
	ex := abbr.ExpanderFor(genre)

	if ex == nil {
		return err
	}

	idString := id.String()

	if idString, err = ex(idString); err != nil {
		err = nil
		return err
	}

	if err = SetWithGenre(id, idString, genre); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func SetWithString(
	id *ObjectId,
	value string,
) (err error) {
	var seq doddish.Seq

	if seq, err = doddish.ScanExactlyOneSeqWithDotAllowedInIdenfierFromString(
		value,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = id.SetWithSeq(seq); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func SetWithGenre(
	id *ObjectId,
	value string,
	genre interfaces.GenreGetter,
) (err error) {
	return id.SetWithGenre(value, genre)

	if err = SetWithString(id, value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	{
		genre := genres.Make(genre.GetGenre())

		if err = genre.AssertGenre(id); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func SetOnlyNotUnknownGenre(
	id *ObjectId,
	value string,
) (err error) {
	if err = SetWithString(id, value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if id.GetGenre() == genres.None {
		err = errors.Errorf("unknown genre for string: %q", value)
		return
	}

	return
}
