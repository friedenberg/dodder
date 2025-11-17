package organize_text

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type Refiner struct {
	Enabled         bool
	UsePrefixJoints bool
}

func (atc *Refiner) shouldMergeAllChildrenIntoParent(a *Assignment) (ok bool) {
	switch {
	case a.Parent.IsRoot:
		fallthrough

	default:
		ok = false
	}

	return ok
}

func (atc *Refiner) shouldMergeIntoParent(a *Assignment) bool {
	ui.Log().Printf("checking node should merge: %s", a)

	if a.Parent == nil {
		ui.Log().Print("parent is nil")
		return false
	}

	if a.Parent.IsRoot {
		ui.Log().Print("parent is root")
		return false
	}

	if a.Transacted.Metadata.GetTags().Len() == 1 && ids.IsEmpty(a.Transacted.Metadata.GetTags().Any()) {
		ui.Log().Print("1 tag, and it's empty, merging")
		return true
	}

	if a.Transacted.Metadata.GetTags().Len() == 0 {
		ui.Log().Print("tags length is 0, merging")
		return true
	}

	if a.Parent.Transacted.Metadata.GetTags().Len() != 1 {
		ui.Log().Print("parent tags length is not 1")
		return false
	}

	if a.Transacted.Metadata.GetTags().Len() != 1 {
		ui.Log().Print("tags length is not 1")
		return false
	}

	equal := quiter.SetEqualsPtr(a.Transacted.Metadata.GetTags(), a.Parent.Transacted.Metadata.GetTags())

	if !equal {
		ui.Log().Print("parent tags not equal")
		return false
	}

	if ids.IsDependentLeaf(a.Parent.Transacted.Metadata.GetTags().Any()) {
		ui.Log().Print("is prefix joint")
		return false
	}

	if ids.IsDependentLeaf(a.Transacted.Metadata.GetTags().Any()) {
		ui.Log().Print("is prefix joint")
		return false
	}

	return true
}

func (atc *Refiner) renameForPrefixJoint(a *Assignment) (err error) {
	if !atc.UsePrefixJoints {
		return err
	}

	if a == nil {
		ui.Log().Printf("assignment is nil")
		return err
	}

	if a.Parent == nil {
		ui.Log().Printf("parent is nil: %#v", a)
		return err
	}

	if a.Parent.Transacted.Metadata.GetTags().Len() == 0 {
		return err
	}

	if a.Parent.Transacted.Metadata.Tags.Len() != 1 {
		return err
	}

	if ids.IsDependentLeaf(a.Parent.Transacted.Metadata.Tags.Any()) {
		return err
	}

	if ids.IsDependentLeaf(a.Transacted.Metadata.Tags.Any()) {
		return err
	}

	if !ids.HasParentPrefix(a.Transacted.Metadata.Tags.Any(), a.Parent.Transacted.Metadata.Tags.Any()) {
		ui.Log().Print("parent is not prefix joint")
		return err
	}

	aEtt := a.Transacted.Metadata.Tags.Any()
	pEtt := a.Parent.Transacted.Metadata.Tags.Any()

	if aEtt.Equals(pEtt) {
		ui.Log().Print("parent is is equal to child")
		return err
	}

	var ls ids.Tag

	if ls, err = ids.LeftSubtract(aEtt, pEtt); err != nil {
		err = errors.Wrap(err)
		return err
	}

	a.Transacted.Metadata.Tags = ids.MakeMutableTagSet(ls)

	return err
}

// passed-in assignment may be nil?
func (atc *Refiner) Refine(a *Assignment) (err error) {
	if !atc.Enabled {
		return err
	}

	if !a.IsRoot {
		if atc.shouldMergeIntoParent(a) {
			ui.Log().Print("merging into parent")
			p := a.Parent

			if err = p.consume(a); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return atc.Refine(p)
		}
	}

	if err = atc.applyPrefixJoints(a); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = atc.renameForPrefixJoint(a); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for _, child := range a.Children {
		if err = atc.Refine(child); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if err = atc.applyPrefixJoints(a); err != nil {
		err = errors.Wrap(err)
		return err
	}

	a.SortChildren()

	return err
}

func (atc Refiner) applyPrefixJoints(a *Assignment) (err error) {
	if !atc.UsePrefixJoints {
		return err
	}

	if a.Transacted.Metadata.Tags == nil || a.Transacted.Metadata.Tags.Len() == 0 {
		return err
	}

	childPrefixes := atc.childPrefixes(a)

	if len(childPrefixes) == 0 {
		return err
	}

	groupingPrefix := childPrefixes[0]

	var na *Assignment

	if a.Transacted.Metadata.Tags.Len() == 1 &&
		a.Transacted.Metadata.Tags.Any().Equals(groupingPrefix.Tag) {
		na = a
	} else {
		na = newAssignment(a.GetDepth() + 1)
		na.Transacted.Metadata.Tags = ids.MakeMutableTagSet(groupingPrefix.Tag)
		a.addChild(na)
	}

	for _, c := range groupingPrefix.assignments {
		if c.Parent != na {
			if err = c.removeFromParent(); err != nil {
				err = errors.Wrap(err)
				return err
			}

			na.addChild(c)
		}

		c.Transacted.Metadata.Tags = ids.SubtractPrefix(
			c.Transacted.Metadata.Tags,
			groupingPrefix.Tag,
		).CloneMutableSetPtrLike()
	}

	return err
}

type tagBag struct {
	ids.Tag
	assignments []*Assignment
}

func (a Refiner) childPrefixes(node *Assignment) (out []tagBag) {
	m := make(map[string][]*Assignment)
	out = make([]tagBag, 0, len(node.Children))

	if node.Transacted.Metadata.Tags.Len() == 0 {
		return out
	}

	for _, c := range node.Children {
		expanded := ids.Expanded(c.Transacted.Metadata.Tags, expansion.ExpanderRight)

		for e := range expanded.All() {
			if e.String() == "" {
				continue
			}

			var n []*Assignment
			ok := false

			if n, ok = m[e.String()]; !ok {
				n = make([]*Assignment, 0)
			}

			n = append(n, c)

			m[e.String()] = n
		}
	}

	for e, n := range m {
		if len(n) > 1 {
			var e1 ids.Tag

			errors.PanicIfError(e1.Set(e))

			out = append(out, tagBag{Tag: e1, assignments: n})
		}
	}

	sort.Slice(
		out,
		func(i, j int) bool {
			if len(out[i].assignments) == len(out[j].assignments) {
				return len(
					out[i].Tag.String(),
				) > len(
					out[j].Tag.String(),
				)
			} else {
				return len(out[i].assignments) > len(out[j].assignments)
			}
		},
	)

	return out
}
