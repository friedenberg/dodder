package ids

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
)

type idExpandable[T any] interface {
	interfaces.ObjectId
	interfaces.GenreGetter
	interfaces.Stringer
}

type idExpandablePtr[T idExpandable[T]] interface {
	interfaces.Ptr[T]
	idExpandable[T]
	interfaces.ObjectId
	interfaces.SetterPtr[T]
}

func expandOne[ID idExpandable[ID], ID_PTR idExpandablePtr[ID]](
	k ID,
	expander expansion.Expander,
	adder interfaces.Adder[ID],
) {
	f := quiter.MakeFuncSetString[ID, ID_PTR](adder)
	expander.Expand(f, k.String())
}

func ExpandOneInto[T interfaces.ObjectId](
	k T,
	mf func(string) (T, error),
	ex expansion.Expander,
	acc interfaces.Adder[T],
) {
	ex.Expand(
		func(v string) (err error) {
			var e T

			if e, err = mf(v); err != nil {
				err = errors.Wrap(err)
				return err
			}

			if err = acc.Add(e); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		},
		k.String(),
	)
}

func ExpandOneSlice[T interfaces.ObjectId](
	k T,
	mf func(string) (T, error),
	exes ...expansion.Expander,
) (out []T) {
	s1 := collections_value.MakeMutableValueSet[T](nil)

	if len(exes) == 0 {
		exes = []expansion.Expander{expansion.ExpanderAll}
	}

	for _, ex := range exes {
		ExpandOneInto(k, mf, ex, s1)
	}

	out = quiter.SortedValuesBy(
		s1,
		func(a, b T) bool {
			return len(a.String()) < len(b.String())
		},
	)

	return out
}

func ExpandMany[ID idExpandable[ID], ID_PTR idExpandablePtr[ID]](
	set interfaces.SetLike[ID],
	expander expansion.Expander,
) (out interfaces.SetPtrLike[ID, ID_PTR]) {
	mutableSet := collections_ptr.MakeMutableValueSetValue[ID, ID_PTR](nil)

	for id := range set.All() {
		expandOne[ID, ID_PTR](id, expander, mutableSet)
	}

	out = mutableSet.CloneSetPtrLike()

	return out
}

func Expanded(s TagSet, ex expansion.Expander) (out TagSet) {
	return ExpandMany(s, ex)
}
