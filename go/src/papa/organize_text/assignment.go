package organize_text

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

// TODO move to object_factory
func newAssignment(depth int) *Assignment {
	assignment := &Assignment{
		Depth:    depth,
		objects:  make(map[string]struct{}),
		Objects:  make(Objects, 0),
		Children: make([]*Assignment, 0),
	}

	sku.TransactedResetter.Reset(&assignment.Transacted)

	return assignment
}

type Assignment struct {
	sku.Transacted

	IsRoot  bool
	Depth   int
	objects map[string]struct{}
	Objects
	Children []*Assignment
	Parent   *Assignment
}

func (assignment *Assignment) AddObject(object *obj) (err error) {
	key := keyer.GetKey(object.sku)
	_, ok := assignment.objects[key]

	if ok {
		return err
	}

	assignment.objects[key] = struct{}{}

	return assignment.Objects.Add(object)
}

func (assignment Assignment) GetDepth() int {
	if assignment.Parent == nil {
		return 0
	} else {
		return assignment.Parent.GetDepth() + 1
	}
}

func (assignment Assignment) MaxDepth() (depth int) {
	depth = assignment.GetDepth()

	for _, child := range assignment.Children {
		childDepth := child.MaxDepth()

		if depth < childDepth {
			depth = childDepth
		}
	}

	return depth
}

func (assignment Assignment) AlignmentSpacing() int {
	if assignment.Transacted.Metadata.GetTags().Len() == 1 && ids.IsDependentLeaf(assignment.Transacted.Metadata.GetTags().Any()) {
		return assignment.Parent.AlignmentSpacing() + len(
			assignment.Parent.Transacted.Metadata.GetTags().Any().String(),
		)
	}

	return 0
}

func (assignment Assignment) MaxLen() (maxLength int) {
	for _, object := range assignment.Objects.All() {
		objectIdLength := object.sku.GetSkuExternal().ObjectId.Len()

		if objectIdLength > maxLength {
			maxLength = objectIdLength
		}
	}

	for _, child := range assignment.Children {
		childMaxLength := child.MaxLen()

		if childMaxLength > maxLength {
			maxLength = childMaxLength
		}
	}

	return maxLength
}

func (assignment Assignment) String() (s string) {
	if assignment.Parent != nil {
		s = assignment.Parent.String() + "."
	}

	return s + quiter.StringCommaSeparated(assignment.Transacted.Metadata.GetTags())
}

func (assignment *Assignment) makeChild(e ids.Tag) (b *Assignment) {
	b = newAssignment(assignment.GetDepth() + 1)
	b.Transacted.GetMetadataMutable().SetTags(ids.MakeMutableTagSet(e))
	assignment.addChild(b)
	return b
}

func (assignment *Assignment) addChild(c *Assignment) {
	if assignment == c {
		panic("child and parent are the same")
	}

	if c.Parent != nil && c.Parent == assignment {
		panic("child already has self as parent")
	}

	if c.Parent != nil {
		panic("child already has a parent")
	}

	assignment.Children = append(assignment.Children, c)
	c.Parent = assignment
}

func (assignment *Assignment) parentOrRoot() (p *Assignment) {
	switch assignment.Parent {
	case nil:
		return assignment

	default:
		return assignment.Parent
	}
}

func (assignment *Assignment) nthParent(n int) (p *Assignment, err error) {
	if n < 0 {
		n = -n
	}

	if n == 0 {
		p = assignment
		return p, err
	}

	if assignment.Parent == nil {
		err = errors.ErrorWithStackf("cannot get nth parent as parent is nil")
		return p, err
	}

	return assignment.Parent.nthParent(n - 1)
}

func (assignment *Assignment) removeFromParent() (err error) {
	return assignment.Parent.removeChild(assignment)
}

func (assignment *Assignment) removeChild(c *Assignment) (err error) {
	if c.Parent != assignment {
		err = errors.ErrorWithStackf("attempting to remove child from wrong parent")
		return err
	}

	if len(assignment.Children) == 0 {
		err = errors.ErrorWithStackf(
			"attempting to remove child when there are no children",
		)
		return err
	}

	cap1 := 0
	cap2 := len(assignment.Children) - 1

	if cap2 > 0 {
		cap1 = cap2
	}

	nc := make([]*Assignment, 0, cap1)

	for _, c1 := range assignment.Children {
		if c1 == c {
			continue
		}

		nc = append(nc, c1)
	}

	c.Parent = nil
	assignment.Children = nc

	return err
}

func (assignment *Assignment) consume(b *Assignment) (err error) {
	for _, c := range b.Children {
		if err = c.removeFromParent(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		assignment.addChild(c)
	}

	for _, obj := range b.Objects.All() {
		assignment.AddObject(obj)
	}
	for _, obj := range b.Objects.All() {
		b.Objects.Del(obj)
	}

	if err = b.removeFromParent(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (assignment *Assignment) AllTags(tags ids.TagSetMutable) (err error) {
	if assignment == nil {
		return err
	}

	var expandedTags ids.TagSet

	if expandedTags, err = assignment.expandedTags(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for tag := range expandedTags.All() {
		if err = tags.Add(tag); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if err = assignment.Parent.AllTags(tags); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (assignment *Assignment) expandedTags() (tags ids.TagSet, err error) {
	tags = ids.MakeTagSet()

	if assignment.Transacted.GetMetadata().GetTags().Len() != 1 || assignment.Parent == nil {
		tags = ids.CloneTagSet(assignment.Transacted.GetMetadata().GetTags())
		return tags, err
	} else {
		e := assignment.Transacted.GetMetadata().GetTags().Any()

		if ids.IsDependentLeaf(e) {
			var pe ids.TagSet

			if pe, err = assignment.Parent.expandedTags(); err != nil {
				err = errors.Wrap(err)
				return tags, err
			}

			if pe.Len() > 1 {
				err = errors.ErrorWithStackf(
					"cannot infer full tag for assignment because parent assignment has more than one tags: %s",
					assignment.Parent.Transacted.GetMetadata().GetTags(),
				)

				return tags, err
			}

			e1 := pe.Any()

			if ids.IsEmpty(e1) {
				err = errors.ErrorWithStackf("parent tag is empty")
				return tags, err
			}

			if err = e.Set(fmt.Sprintf("%s%s", e1, e)); err != nil {
				err = errors.Wrap(err)
				return tags, err
			}
		}

		tags = ids.MakeTagSet(e)
	}

	return tags, err
}

func (assignment *Assignment) SubtractFromSet(
	tagsToSubtract ids.TagSetMutable,
) (err error) {
	for assignmentTag := range assignment.Transacted.GetMetadata().AllTags() {
		for tagToSubtract := range tagsToSubtract.All() {
			if ids.ContainsExactly(tagToSubtract, assignmentTag) {
				if err = tagsToSubtract.Del(tagToSubtract); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}
		}

		if err = tagsToSubtract.Del(assignmentTag); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if assignment.Parent == nil {
		return err
	}

	return assignment.Parent.SubtractFromSet(tagsToSubtract)
}

func (assignment *Assignment) Contains(e *ids.Tag) bool {
	if assignment.Transacted.GetMetadata().GetTags().ContainsKey(e.String()) {
		return true
	}

	if assignment.Parent == nil {
		return false
	}

	return assignment.Parent.Contains(e)
}

func (assignment *Assignment) SortChildren() {
	sort.Slice(assignment.Children, func(i, j int) bool {
		esi := assignment.Children[i].Transacted.GetMetadata().GetTags()
		esj := assignment.Children[j].Transacted.GetMetadata().GetTags()

		if esi.Len() == 1 && esj.Len() == 1 {
			ei := strings.TrimPrefix(esi.Any().String(), "-")
			ej := strings.TrimPrefix(esj.Any().String(), "-")

			ii, ierr := strconv.ParseInt(ei, 0, 64)
			ij, jerr := strconv.ParseInt(ej, 0, 64)

			if ierr == nil && jerr == nil {
				return ii < ij
			} else {
				return ei < ej
			}
		} else {
			vi := quiter.StringCommaSeparated(esi)
			vj := quiter.StringCommaSeparated(esj)
			return vi < vj
		}
	})
}
