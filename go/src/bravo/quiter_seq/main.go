package quiter_seq

import "code.linenisgreat.com/dodder/go/src/_/interfaces"

func Any[ELEMENT any](seq interfaces.Seq[ELEMENT]) (element ELEMENT) {
	for element = range seq {
		break
	}

	return
}

func SeqWithIndex[ELEMENT any](seq interfaces.Seq[ELEMENT]) interfaces.Seq2[int, ELEMENT] {
	return func(yield func(int, ELEMENT) bool) {
		var index int

		for element := range seq {
			if !yield(index, element) {
				return
			}

			index++
		}
	}
}
