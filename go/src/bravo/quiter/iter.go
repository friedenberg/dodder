package quiter

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MakeSeqErrorFromSeq[ELEMENT any](
	iter interfaces.Seq[ELEMENT],
) interfaces.SeqError[ELEMENT] {
	return func(yield func(ELEMENT, error) bool) {
		for element := range iter {
			if !yield(element, nil) {
				break
			}
		}
	}
}
