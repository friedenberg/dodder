package ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
)

func ExpandTagSet(set TagSet, expander expansion.Expander) TagSetMutable {
	setMutable := MakeTagSetMutable()

	for tag := range expansion.ExpandMany(set.All(), expander) {
		setMutable.Add(tag)
	}

	return setMutable
}
