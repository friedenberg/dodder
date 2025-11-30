package ids

import (
	"fmt"
	"slices"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/alfa/quiter_seq"
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

func ExpandOneIntoIds[
	ID idExpandable[ID],
	ID_PTR idExpandablePtr[ID],
](
	identifierString string,
	expander expansion.Expander,
) interfaces.SeqError[ID] {
	return func(yield func(ID, error) bool) {
		if identifierString == "" {
			return
		}

		var expandedId ID

		for expanded := range expander.Expand(identifierString) {
			if expanded == "" {
				continue

				// TODO move this check into expansion.*
				panic(
					fmt.Sprintf(
						"empty expansion for original identifier %q for expander %T",
						identifierString,
						expander,
					),
				)
			}

			if err := ID_PTR(&expandedId).Set(expanded); err != nil {
				if !yield(expandedId, err) {
					return
				}
			}

			if !yield(expandedId, nil) {
				return
			}
		}
	}
}

func ExpandIntoSlice[
	ID interfaces.ObjectId,
	ID_PTR idExpandablePtr[ID],
](
	token string,
	expander expansion.Expander,
) collections_slice.Slice[ID] {
	return slices.Collect(
		quiter_seq.SeqErrorToSeqAndPanic(
			ExpandOneIntoIds[ID, ID_PTR](
				token,
				expander,
			),
		),
	)
}

// TODO remove
func ExpandOneInto[
	ID interfaces.ObjectId,
	ID_PTR idExpandablePtr[ID],
](
	token ID,
	funcSetString func(string) (ID, error),
	expander expansion.Expander,
	adder interfaces.Adder[ID],
) {
	for id := range ExpandOneIntoIds[ID, ID_PTR](token.String(), expander) {
		if err := adder.Add(id); err != nil {
			panic(err)
		}
	}
}

func ExpandMany[ID idExpandable[ID], ID_PTR idExpandablePtr[ID]](
	seq interfaces.Seq[ID],
	expander expansion.Expander,
) interfaces.Seq[ID] {
	return func(yield func(ID) bool) {
		for id := range seq {
			seqExpansion := quiter_seq.SeqErrorToSeqAndPanic(
				ExpandOneIntoIds[ID, ID_PTR](id.String(), expander),
			)

			for expanded := range seqExpansion {
				if !yield(expanded) {
					return
				}
			}
		}
	}
}

func ExpandTagSet(set TagSet, expander expansion.Expander) TagSetMutable {
	setMutable := MakeTagSetMutable()

	for tag := range ExpandMany(set.All(), expander) {
		setMutable.Add(tag)
	}

	return setMutable
}
