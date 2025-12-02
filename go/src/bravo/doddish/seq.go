package doddish

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
)

type Seq collections_slice.Slice[Token]

func (seq Seq) Len() int {
	return len(seq)
}

func (seq Seq) At(idx int) Token {
	return seq[idx]
}

func (seq *Seq) GetSlice() collections_slice.Slice[Token] {
	return (collections_slice.Slice[Token])(*seq)
}

func (seq *Seq) GetSliceMutable() *collections_slice.Slice[Token] {
	return (*collections_slice.Slice[Token])(seq)
}

func (seq *Seq) Add(tokenType TokenType, contents []byte) {
	seq.GetSliceMutable().Append(Token{TokenType: tokenType, Contents: contents})
}

func (seq Seq) StringDebug() string {
	var sb strings.Builder

	sb.WriteString("Seq{")
	for _, t := range seq {
		fmt.Fprintf(&sb, "%s:%q ", t.TokenType, t.Contents)
	}
	sb.WriteString("}")

	return sb.String()
}

func (seq Seq) String() string {
	var sb strings.Builder

	for _, t := range seq {
		sb.Write(t.Contents)
	}

	return sb.String()
}

func (seq Seq) Clone() (dst Seq) {
	dst = make(Seq, len(seq))

	for i := range seq {
		dst[i] = seq[i].Clone()
	}

	return dst
}

func (seq *Seq) Reset() {
	seq.GetSliceMutable().Reset()
}

func (seq Seq) GetTokenTypes() TokenTypes {
	out := make(TokenTypes, seq.Len())

	for i := range out {
		out[i] = seq.At(i).TokenType
	}

	return out
}

func (seq Seq) MatchAll(tokens ...TokenMatcher) bool {
	if len(tokens) != seq.Len() {
		return false
	}

	for i, m := range tokens {
		if !m.Match(seq.At(i)) {
			return false
		}
	}

	return true
}

func (seq Seq) MatchStart(tokens ...TokenMatcher) bool {
	if len(tokens) > seq.Len() {
		return false
	}

	for i, m := range tokens {
		if !m.Match(seq.At(i)) {
			return false
		}
	}

	return true
}

func (seq Seq) MatchEnd(tokens ...TokenMatcher) (ok bool, left, right Seq) {
	if len(tokens) > seq.Len() {
		return ok, left, right
	}

	for i := seq.Len() - 1; i >= 0; i-- {
		partition := seq.At(i)
		j := len(tokens) - (seq.Len() - i)

		if j < 0 {
			break
		}

		m := tokens[j]

		if !m.Match(partition) {
			return ok, left, right
		}

		left = seq[:i]
		right = seq[i:]
	}

	ok = true

	return ok, left, right
}

func (seq Seq) PartitionFavoringRight(
	m TokenMatcher,
) (ok bool, left, right Seq, partition Token) {
	for i := seq.Len() - 1; i >= 0; i-- {
		partition = seq.At(i)

		if m.Match(partition) {
			ok = true
			left = seq[:i]
			right = seq[i+1:]
			return ok, left, right, partition
		}
	}

	return ok, left, right, partition
}

func (seq Seq) PartitionFavoringLeft(
	m TokenMatcher,
) (ok bool, left, right Seq, partition Token) {
	for i := range seq {
		partition = seq.At(i)

		if m.Match(partition) {
			ok = true
			left = seq[:i]
			right = seq[i+1:]
			return ok, left, right, partition
		}
	}

	return ok, left, right, partition
}
