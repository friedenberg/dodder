package alfred

import (
	"bytes"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var poolMatchBuilder interfaces.Pool[MatchBuilder, *MatchBuilder]

func init() {
	poolMatchBuilder = pool.MakePool[MatchBuilder, *MatchBuilder](
		NewMatchBuilder,
		func(mb *MatchBuilder) {
			mb.Buffer.Reset()
		},
	)
}

func GetPoolMatchBuilder() interfaces.Pool[MatchBuilder, *MatchBuilder] {
	return poolMatchBuilder
}

type MatchBuilder struct {
	bytes.Buffer
}

func NewMatchBuilder() *MatchBuilder {
	return &MatchBuilder{}
}

var sliceBytesUnderscore = []byte("_")

func (mb *MatchBuilder) AddMatchBytes(s []byte) {
	s1 := bytes.Split(s, sliceBytesUnderscore)

	for _, s2 := range s1 {
		mb.Write(s2)
		mb.WriteRune(' ')
	}
}

func (mb *MatchBuilder) AddMatch(s string) {
	s1 := strings.Split(s, "_")

	for _, s2 := range s1 {
		mb.WriteString(s2)
		mb.WriteString(" ")
	}
}

func (mb *MatchBuilder) AddMatches(s ...string) {
	for _, v := range s {
		mb.AddMatch(v)
	}
}

func (mb *MatchBuilder) Bytes() []byte {
	return mb.Buffer.Bytes()
}
