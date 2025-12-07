package doddish

import (
	"unicode/utf8"

	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
)

func SeqCompare(left, right Seq) cmp.Result {
	return seqCompare(left, right, false)
}

func SeqComparePartial(left, right Seq) cmp.Result {
	return seqCompare(left, right, true)
}

func seqCompare(left, right Seq, partial bool) cmp.Result {
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
			partial,
		)

		if result.IsEqual() {
			continue
		} else {
			return result
		}
	}
}

type ComparableSeq struct {
	Tokens         []Token
	TokenIndex     int
	TokenByteIndex int
	ByteCount      int
}

func (seq Seq) GetComparable() ComparableSeq {
	var byteCount int

	for _, token := range seq {
		byteCount += len(token.Contents)
	}

	return ComparableSeq{
		Tokens:    seq,
		ByteCount: byteCount,
	}
}

func (seq ComparableSeq) Len() int {
	return seq.ByteCount
}

func (seq ComparableSeq) getRemainingTokenBytes() []byte {
	return seq.Tokens[seq.TokenIndex].Contents[seq.TokenByteIndex:]
}

func (seq ComparableSeq) DecodeRune() (char rune, width int) {
	char, width = utf8.DecodeRune(seq.getRemainingTokenBytes())
	return char, width
}

func (seq ComparableSeq) Shift(amount int) ComparableSeq {
	remainingBytes := seq.getRemainingTokenBytes()
	seq.ByteCount -= amount

	if len(remainingBytes) == amount {
		seq.TokenByteIndex = 0
		seq.TokenIndex++
	} else {
		seq.TokenByteIndex += amount
	}

	return seq
}
