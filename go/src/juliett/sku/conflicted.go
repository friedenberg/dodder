package sku

import (
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

func (c *Conflicted) FindBestCommonAncestor(
	negotiator ParentNegotiator,
) (err error) {
	if negotiator == nil {
		return
	}

	if c.Base, err = negotiator.FindBestCommonAncestor(*c); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (c Conflicted) GetCollection() Collection {
	return c
}

func (c Conflicted) Len() int {
	if c.Base == nil {
		return 2
	} else {
		return 3
	}
}

func (c Conflicted) Any() *Transacted {
	return c.Local
}

func (c Conflicted) All() interfaces.Seq[*Transacted] {
	return func(yield func(*Transacted) bool) {
		if !yield(c.Local) {
			return
		}

		if c.Base != nil && !yield(c.Base) {
			return
		}

		if !yield(c.Remote) {
			return
		}
	}
}

func (c Conflicted) IsAllInlineType(itc ids.InlineTypeChecker) bool {
	if !itc.IsInlineType(c.Local.GetType()) {
		return false
	}

	if c.Base != nil && !itc.IsInlineType(c.Base.GetType()) {
		return false
	}

	if !itc.IsInlineType(c.Remote.GetType()) {
		return false
	}

	return true
}

func (tm *Conflicted) MergeTags() (err error) {
	if tm.Base == nil {
		return
	}

	left := tm.Local.GetTags().CloneMutableSetPtrLike()
	middle := tm.Base.GetTags().CloneMutableSetPtrLike()
	right := tm.Remote.GetTags().CloneMutableSetPtrLike()

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

	if err = middle.EachPtr(
		func(e *ids.Tag) (err error) {
			if left.ContainsKey(left.KeyPtr(e)) && right.ContainsKey(right.KeyPtr(e)) {
				return removeFromAllButAddTo(e, same)
			} else if left.ContainsKey(left.KeyPtr(e)) || right.ContainsKey(right.KeyPtr(e)) {
				return removeFromAllButAddTo(e, deleted)
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = left.EachPtr(same.AddPtr); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = right.EachPtr(same.AddPtr); err != nil {
		err = errors.Wrap(err)
		return
	}

	ets := same.CloneSetPtrLike()

	tm.Local.GetMetadata().SetTags(ets)
	tm.Base.GetMetadata().SetTags(ets)
	tm.Remote.GetMetadata().SetTags(ets)

	return
}

func (c *Conflicted) ReadConflictMarker(
	iter interfaces.SeqError[*Transacted],
) (err error) {
	i := 0

	for sk, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		switch i {
		case 0:
			c.Local = sk

		case 1:
			c.Base = sk

		case 2:
			c.Remote = sk

		default:
			err = errors.ErrorWithStackf("too many skus in conflict file")
			return
		}

		i++
	}

	// Conflicts can exist between objects without a base
	if i == 2 {
		c.Remote = c.Base
		c.Base = nil
	}

	return
}
