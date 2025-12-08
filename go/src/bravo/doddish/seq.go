package doddish

import (
	"fmt"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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
	seq.GetSliceMutable().Append(Token{Type: tokenType, Contents: contents})
}

func (seq Seq) StringDebug() string {
	var sb strings.Builder

	sb.WriteString("Seq{")
	for _, t := range seq {
		fmt.Fprintf(&sb, "%s:%q ", t.Type, t.Contents)
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
		out[i] = seq.At(i).Type
	}

	return out
}

func (seq Seq) GetBinaryMarshaler() SeqBinaryMarshaler {
	return (SeqBinaryMarshaler)(seq)
}

func (seq *Seq) GetBinaryUnmarshaler() *SeqBinaryMarshaler {
	return (*SeqBinaryMarshaler)(seq)
}

type SeqBinaryMarshaler Seq

func (marshaler SeqBinaryMarshaler) ToSeq() Seq {
	return Seq(marshaler)
}

func (marshaler *SeqBinaryMarshaler) ToSeqMutable() *Seq {
	return (*Seq)(marshaler)
}

func (marshaler SeqBinaryMarshaler) MarshalBinary() (bites []byte, err error) {
	return marshaler.AppendBinary(nil)
}

func (marshaler SeqBinaryMarshaler) AppendBinary(bites []byte) ([]byte, error) {
	var byteCount int

	for _, token := range marshaler.ToSeq() {
		byteCount += len(token.Contents) + 1 + 1
	}

	bites = slices.Grow(bites, byteCount)

	for _, token := range marshaler.ToSeq() {
		bites = append(bites, byte(len(token.Contents)+1))

		var err error

		if bites, err = token.AppendBinary(bites); err != nil {
			err = errors.Wrap(err)
			return bites, err
		}
	}

	return bites, nil
}

func (marshaler *SeqBinaryMarshaler) UnmarshalBinary(bites []byte) (err error) {
	for len(bites) > 0 {
		byteCount := int(bites[0])
		next := bites[1:byteCount]
		bites = bites[byteCount+1:]

		var token Token

		if err = token.UnmarshalBinary(next); err != nil {
			err = errors.Wrap(err)
			return err
		}

		marshaler.ToSeqMutable().GetSliceMutable().Append(token)
	}

	return
}
