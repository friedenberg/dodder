package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
)

type (
	TagStruct  = tagStruct
	TypeStruct = typeStruct

	Id interface {
		interfaces.ObjectId
		ToSeq() doddish.Seq
		ToType() TypeStruct
		IsEmpty() bool
	}

	Tag   = Id
	SeqId = objectId3
	Type  = Id

	// TODO rename to BinaryTypeChecker and flip uses
	InlineTypeChecker interface {
		IsInlineType(Type) bool
	}

	// TODO remove
	IdWithParts interface {
		Id
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

func MustObjectId(idWithParts IdWithParts) (id *ObjectId) {
	id = &ObjectId{}
	err := id.SetWithIdLike(idWithParts)
	errors.PanicIfError(err)
	return id
}

func MakeObjectId(value string) (objectId *ObjectId, err error) {
	var boxScanner doddish.Scanner
	reader, repool := pool.GetStringReader(value)
	defer repool()
	boxScanner.Reset(reader)

	objectId = GetObjectIdPool().Get()

	if value == "" {
		return objectId, err
	}

	if !boxScanner.ScanDotAllowedInIdentifiers() {
		return objectId, err
	}

	seq := boxScanner.GetSeq()

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
