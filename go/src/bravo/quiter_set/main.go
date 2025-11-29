package quiter_set

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_seq"
)

func Any[ELEMENT any](set interfaces.SetBase[ELEMENT]) (element ELEMENT) {
	return quiter_seq.Any(set.All())
}

func AllWithIndex[ELEMENT any](set interfaces.SetBase[ELEMENT]) interfaces.Seq2[int, ELEMENT] {
	return quiter_seq.SeqWithIndex(set.All())
}
