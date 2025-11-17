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

func expandOne[T idExpandable[T], TPtr idExpandablePtr[T]](
	k TPtr,
	ex expansion.Expander,
	acc interfaces.Adder[T],
) {
	f := quiter.MakeFuncSetString[T, TPtr](acc)
	ex.Expand(f, k.String())
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

func ExpandMany[T idExpandable[T], TPtr idExpandablePtr[T]](
	ks interfaces.SetPtrLike[T, TPtr],
	ex expansion.Expander,
) (out interfaces.SetPtrLike[T, TPtr]) {
	s1 := collections_ptr.MakeMutableValueSetValue[T, TPtr](nil)

	for k := range ks.AllPtr() {
		expandOne[T, TPtr](k, ex, s1)
	}

	out = s1.CloneSetPtrLike()

	return out
}

func Expanded(s TagSet, ex expansion.Expander) (out TagSet) {
	return ExpandMany(s, ex)
}
