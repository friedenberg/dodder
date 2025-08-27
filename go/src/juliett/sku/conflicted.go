package sku

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type ParentNegotiator interface {
	FindBestCommonAncestor(Conflicted) (*Transacted, error)
}

// TODO consider making this a ConflictedWithBase and ConflictedWithoutBase
// and an interface for both
type Conflicted struct {
	*CheckedOut
	Local, Base, Remote *Transacted
}

func (conflicted Conflicted) String() string {
	return fmt.Sprintf(
		"Local: %q\nBase: %q\nRemote: %q\n",
		String(conflicted.Local),
		String(conflicted.Base),
		String(conflicted.Remote),
	)
}

func (conflicted *Conflicted) FindBestCommonAncestor(
	negotiator ParentNegotiator,
) (err error) {
	if negotiator == nil {
		return
	}

	if conflicted.Base, err = negotiator.FindBestCommonAncestor(*conflicted); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (conflicted Conflicted) GetCollection() Collection {
	return conflicted
}

func (conflicted Conflicted) Len() int {
	if conflicted.Base == nil {
		return 2
	} else {
		return 3
	}
}

func (conflicted Conflicted) Any() *Transacted {
	return conflicted.Local
}

func (conflicted Conflicted) All() interfaces.Seq[*Transacted] {
	return func(yield func(*Transacted) bool) {
		if !yield(conflicted.Local) {
			return
		}

		if conflicted.Base != nil && !yield(conflicted.Base) {
			return
		}

		if !yield(conflicted.Remote) {
			return
		}
	}
}

func (conflicted Conflicted) IsAllInlineType(
	typeChecker ids.InlineTypeChecker,
) bool {
	if typeChecker == nil {
		panic("nil type checker")
	}

	if conflicted.Local == nil {
		panic("nil local")
	}

	if !typeChecker.IsInlineType(conflicted.Local.GetType()) {
		return false
	}

	if conflicted.Base != nil &&
		!typeChecker.IsInlineType(conflicted.Base.GetType()) {
		return false
	}

	if !typeChecker.IsInlineType(conflicted.Remote.GetType()) {
		return false
	}

	return true
}

func (conflicted *Conflicted) MergeTags() (err error) {
	if conflicted.Base == nil {
		return
	}

	left := conflicted.Local.GetTags().CloneMutableSetPtrLike()
	middle := conflicted.Base.GetTags().CloneMutableSetPtrLike()
	right := conflicted.Remote.GetTags().CloneMutableSetPtrLike()

	same := ids.MakeTagMutableSet()
	deleted := ids.MakeTagMutableSet()

	removeFromAllButAddTo := func(
		e *ids.Tag,
		toAdd ids.TagMutableSet,
	) (err error) {
		if err = toAdd.AddPtr(e); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = left.DelPtr(e); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = middle.DelPtr(e); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = right.DelPtr(e); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}

	for tag := range middle.AllPtr() {
		if left.ContainsKey(left.KeyPtr(tag)) &&
			right.ContainsKey(right.KeyPtr(tag)) {
			if err = removeFromAllButAddTo(tag, same); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else if left.ContainsKey(left.KeyPtr(tag)) || right.ContainsKey(right.KeyPtr(tag)) {
			if err = removeFromAllButAddTo(tag, deleted); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	for tag := range left.AllPtr() {
		if err = same.AddPtr(tag); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	for tag := range right.AllPtr() {
		if err = same.AddPtr(tag); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	tags := same.CloneSetPtrLike()

	conflicted.Local.GetMetadata().SetTags(tags)
	conflicted.Base.GetMetadata().SetTags(tags)
	conflicted.Remote.GetMetadata().SetTags(tags)

	return
}

func (conflicted *Conflicted) ReadConflictMarker(
	seq interfaces.SeqError[*Transacted],
) (err error) {
	idx := 0

	for object, iterErr := range seq {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		switch idx {
		case 0:
			conflicted.Local = object

		case 1:
			conflicted.Base = object

		case 2:
			conflicted.Remote = object

		default:
			err = errors.ErrorWithStackf("too many objects in conflict file")
			return
		}

		idx++
	}

	if idx == 0 {
		err = errors.ErrorWithStackf("no objects in conflict file")
		return
	}

	// Conflicts can exist between objects without a base
	if idx == 2 {
		conflicted.Remote = conflicted.Base
		conflicted.Base = nil
	}

	return
}
