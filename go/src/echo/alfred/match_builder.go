package alfred

import (
	"bytes"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var poolMatchBuilder = pool.MakePool(
	NewMatchBuilder,
	func(matchBuilder *MatchBuilder) {
		matchBuilder.Buffer.Reset()
	},
)

// TODO switch to returning repool function
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

func (matchBuilder *MatchBuilder) AddMatchBytes(s []byte) {
	s1 := bytes.SplitSeq(s, sliceBytesUnderscore)

	for s2 := range s1 {
		matchBuilder.Write(s2)
		matchBuilder.WriteRune(' ')
	}
}

func (matchBuilder *MatchBuilder) AddMatch(s string) {
	s1 := strings.SplitSeq(s, "_")

	for s2 := range s1 {
		matchBuilder.WriteString(s2)
		matchBuilder.WriteString(" ")
	}
}

func (matchBuilder *MatchBuilder) AddMatches(s ...string) {
	for _, v := range s {
		matchBuilder.AddMatch(v)
	}
}

func (matchBuilder *MatchBuilder) Bytes() []byte {
	return matchBuilder.Buffer.Bytes()
}
