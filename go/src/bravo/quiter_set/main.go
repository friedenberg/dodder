package quiter_set

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_seq"
)

func Any[ELEMENT any](set interfaces.Set[ELEMENT]) (element ELEMENT) {
	return quiter_seq.Any(set.All())
}

func AllWithIndex[ELEMENT any](set interfaces.Set[ELEMENT]) interfaces.Seq2[int, ELEMENT] {
	return quiter_seq.SeqWithIndex(set.All())
}

func Contains[ELEMENT any](
	set interfaces.Set[ELEMENT],
	element ELEMENT,
) bool {
	return set.ContainsKey(set.Key(element))
}

func Del[ELEMENT any](set interfaces.MutableSetLike[ELEMENT], element ELEMENT) {
	key := set.Key(element)
	errors.PanicIfError(set.DelKey(key))
}
