package doddish

import (
	"encoding"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Token struct {
	Type     TokenType
	Contents []byte
}

var (
	_ encoding.BinaryMarshaler   = Token{}
	_ encoding.BinaryAppender    = Token{}
	_ encoding.BinaryUnmarshaler = &Token{}
)

func (token Token) String() string {
	return string(token.Contents)
}

func (token Token) Clone() (dst Token) {
	dst.Type = token.Type
	dst.Contents = slices.Clone(token.Contents)
	return dst
}

func (token Token) MarshalBinary() ([]byte, error) {
	return token.AppendBinary(nil)
}

func (token Token) AppendBinary(bites []byte) ([]byte, error) {
	bites = slices.Grow(bites, len(token.Contents)+1)
	bites = append(bites, byte(token.Type))
	bites = append(bites, token.Contents...)
	return bites, nil
}

func (token *Token) UnmarshalBinary(bites []byte) (err error) {
	if len(bites) < 2 {
		err = errors.Errorf("expected at least two bytes but got %x", bites)
		return err
	}

	token.Type = TokenType(bites[0])
	token.Contents = bites[1:]

	return err
}
