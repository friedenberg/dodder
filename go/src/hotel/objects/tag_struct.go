package objects

import "code.linenisgreat.com/dodder/go/src/alfa/cmp"

type tagStruct struct {
	// TODO add path information

	// required to be exported for Gob's stupid illusions
	Lock tagLockStruct
}

func tagStructCompareTagKey(left, right tagStruct) cmp.Result {
	return cmp.String(
		left.Lock.GetKey().String(),
		right.Lock.GetKey().String(),
	)
}
