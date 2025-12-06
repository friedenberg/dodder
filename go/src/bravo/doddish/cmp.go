package doddish

import (
	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
)

func SeqCompare(left, right Seq) cmp.Result {
	lenLeft, lenRight := left.Len(), right.Len()

	// TODO remove?
	switch {
	case lenLeft == 0 && lenRight == 0:
		return cmp.Equal

	case lenLeft == 0:
		return cmp.Less

	case lenRight == 0:
		return cmp.Greater
	}

	for {
		lenLeft, lenRight := left.Len(), right.Len()

		switch {
		case lenLeft == 0 && lenRight == 0:
			return cmp.Equal

		case lenLeft == 0:
			if lenRight <= lenLeft {
				return cmp.Equal
			} else {
				return cmp.Less
			}

		case lenRight == 0:
			return cmp.Greater
		}

		tokenLeft := left.GetSlice().First()
		tokenRight := right.GetSlice().First()

		result := cmp.CompareUTF8Bytes(
			tokenLeft.Contents,
			tokenRight.Contents,
			false,
		)

		if result.IsEqual() {
			continue
		} else {
			return result
		}
	}
}
