package box

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=SeqType
type SeqType int

const (
	SeqTypeUnknown    = SeqType(iota)
	SeqTypeIdentifier // one/uno, tag, !type, /browser/bookmark-1, @abcd
	SeqTypeField      // url=blah blob=hello contents="wow" contents="wow with\" quote"
)

type Seq []Token

func (seq Seq) Len() int {
	return len(seq)
}

func (seq Seq) At(idx int) Token {
	return seq[idx]
}

func (a Seq) EqualsSeq(b Seq) bool {
	return false
}

func (seq *Seq) Add(tokenType TokenType, contents []byte) {
	*seq = append(*seq, Token{TokenType: tokenType, Contents: contents})
}

func (seq *Seq) AddToken(token Token) {
	*seq = append(*seq, token)
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

func (src Seq) Clone() (dst Seq) {
	dst = make(Seq, len(src))

	for i := range src {
		dst[i] = src[i].Clone()
	}

	return
}

func (seq *Seq) Reset() {
	*seq = Seq([]Token{})
}

type TokenTypes []TokenType

func (actual TokenTypes) Equals(expected ...TokenType) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i, a := range actual {
		if a != expected[i] {
			return false
		}
	}

	return true
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
		return
	}

	for i := seq.Len() - 1; i >= 0; i-- {
		partition := seq.At(i)
		j := len(tokens) - (seq.Len() - i)

		if j < 0 {
			break
		}

		m := tokens[j]

		if !m.Match(partition) {
			return
		}

		left = seq[:i]
		right = seq[i:]
	}

	ok = true

	return
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
			return
		}
	}

	return
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
			return
		}
	}

	return
}
