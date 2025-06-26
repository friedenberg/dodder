package box

import (
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

type testToken struct {
	TokenType
	Contents string
}

func (token testToken) String() string {
	return fmt.Sprintf("%s %s", token.TokenType, token.Contents)
}

func makeTestToken(tt TokenType, contents string) testToken {
	return testToken{
		TokenType: tt,
		Contents:  contents,
	}
}

type testSeq []testToken

func makeTestSeq(tokens ...any) (ts testSeq) {
	for i := 0; i < len(tokens); i += 2 {
		ts = append(ts,
			makeTestToken(
				tokens[i].(TokenType),
				tokens[i+1].(string),
			),
		)
	}

	return
}

func makeTestSeqFromSeq(seq Seq) (ts testSeq) {
	for _, t := range seq {
		ts = append(ts, testToken{
			TokenType: t.TokenType,
			Contents:  string(t.Contents),
		})
	}

	return
}

func makeSeqFromTestSeq(seq testSeq) (ts Seq) {
	for _, t := range seq {
		ts = append(ts, Token{
			TokenType: t.TokenType,
			Contents:  []byte(t.Contents),
		})
	}

	return
}
