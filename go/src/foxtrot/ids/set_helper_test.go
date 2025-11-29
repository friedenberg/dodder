package ids

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func assertSetRemovesPrefixes(
	t1 *testing.T,
	ac1 TagSet,
	ex TagSet,
	prefix string,
) {
	t := &ui.T{T: t1}
	t = t.Skip(1)

	ac := CloneTagSetMutable(ac1)
	RemovePrefixes(ac, MustTag(prefix))

	if !quiter_set.Equals(ac, ex) {
		t.Errorf(
			"removing prefixes doesn't match:\nexpected: %q\n  actual: %q",
			ex,
			ac,
		)
	}
}
