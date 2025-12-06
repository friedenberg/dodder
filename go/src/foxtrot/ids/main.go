package ids

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
)

type Id interface {
	interfaces.ObjectId
	ToSeq() SeqId
}

func MakeObjectId(value string) (objectId *ObjectId, err error) {
	var boxScanner doddish.Scanner
	reader, repool := pool.GetStringReader(value)
	defer repool()
	boxScanner.Reset(reader)

	objectId = &ObjectId{}

	if value == "" {
		return objectId, err
	}

	if !boxScanner.ScanDotAllowedInIdentifiers() {
		return objectId, err
	}

	seq := boxScanner.GetSeq()

	if err = objectId.ReadFromSeq(seq); err != nil {
		return objectId, err
	}

	return objectId, err
}

func Equals(left, right interfaces.ObjectId) (ok bool) {
	if left.GetGenre().String() != right.GetGenre().String() {
		return ok
	}

	if left.String() != right.String() {
		return ok
	}

	return true
}

func FormattedString(k interfaces.ObjectIdWithParts) string {
	sb := &strings.Builder{}
	parts := k.Parts()
	sb.WriteString(parts[0])
	sb.WriteString(parts[1])
	sb.WriteString(parts[2])
	return sb.String()
}

func AlignedParts(
	id interfaces.ObjectIdWithParts, lenLeft, lenRight int,
) (string, string, string) {
	parts := id.Parts()
	left := parts[0]
	middle := parts[1]
	right := parts[2]

	diffLeft := lenLeft - len(left)
	if diffLeft > 0 {
		left = strings.Repeat(" ", diffLeft) + left
	}

	diffRight := lenRight - len(right)
	if diffRight > 0 {
		right = right + strings.Repeat(" ", diffRight)
	}

	return left, middle, right
}

func Aligned(id interfaces.ObjectIdWithParts, lenLeft, lenRight int) string {
	left, middle, right := AlignedParts(id, lenLeft, lenRight)
	return fmt.Sprintf("%s%s%s", left, middle, right)
}

func LeftSubtract[
	T interfaces.Stringer,
	TPtr interfaces.StringSetterPtr[T],
](
	a, b T,
) (c T, err error) {
	if err = TPtr(&c).Set(strings.TrimPrefix(a.String(), b.String())); err != nil {
		err = errors.Wrapf(err, "'%s' - '%s'", a, b)
		return c, err
	}

	return c, err
}

func Contains(a, b interfaces.ObjectIdWithParts) bool {
	var (
		as = a.Parts()
		bs = b.Parts()
	)

	for i, e := range as {
		if !strings.HasPrefix(e, bs[i]) {
			return false
		}
	}

	return true
}

func ContainsExactly(a, b interfaces.ObjectIdWithParts) bool {
	var (
		as = a.Parts()
		bs = b.Parts()
	)

	if as != bs {
		return false
	}

	return true
}

func IsEmpty[T interfaces.Stringer](a T) bool {
	return len(a.String()) == 0
}
