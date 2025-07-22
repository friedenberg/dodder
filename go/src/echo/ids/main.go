package ids

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/box"
)

type Abbreviatable interface {
	interfaces.Stringer
}

type (
	ObjectIdLike interface {
		interfaces.GenreGetter
		interfaces.Stringer
		IsEmpty() bool
		SetObjectIdLike(ObjectIdLike) error
	}

	ExternalObjectIdGetter interface {
		GetExternalObjectId() ExternalObjectIdLike
	}

	ExternalObjectIdLike interface {
		ObjectIdLike
		ExternalObjectIdGetter
	}
)

type Index struct{}

func MakeObjectId(v string) (objectId *ObjectId, err error) {
	var boxScanner box.Scanner
	reader, repool := pool.GetStringReader(v)
	defer repool()
	boxScanner.Reset(reader)

	objectId = &ObjectId{}

	if v == "" {
		return
	}

	if !boxScanner.ScanDotAllowedInIdentifiers() {
		return
	}

	seq := boxScanner.GetSeq()

	if err = objectId.ReadFromSeq(seq); err != nil {
		return
	}

	return
}

func Equals(a, b interfaces.ObjectId) (ok bool) {
	if a.GetGenre().GetGenreString() != b.GetGenre().GetGenreString() {
		return
	}

	if a.String() != b.String() {
		return
	}

	return true
}

func FormattedString(k interfaces.ObjectId) string {
	sb := &strings.Builder{}
	parts := k.Parts()
	sb.WriteString(parts[0])
	sb.WriteString(parts[1])
	sb.WriteString(parts[2])
	return sb.String()
}

func AlignedParts(
	id interfaces.ObjectId, lenLeft, lenRight int,
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

func Aligned(id interfaces.ObjectId, lenLeft, lenRight int) string {
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
		return
	}

	return
}

func Contains(a, b interfaces.ObjectId) bool {
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

func ContainsExactly(a, b interfaces.ObjectId) bool {
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
