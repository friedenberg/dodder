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
func newAssignment(d int) *Assignment {
	a := &Assignment{
		Depth:    d,
		objects:  make(map[string]struct{}),
		Objects:  make(Objects, 0),
		Children: make([]*Assignment, 0),
	}

	sku.TransactedResetter.Reset(&a.Transacted)

	return a
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

func (a *Assignment) AddObject(v *obj) (err error) {
	k := keyer.GetKey(v.sku)
	_, ok := a.objects[k]

	if ok {
		return err
	}

	a.objects[k] = struct{}{}

	return a.Objects.Add(v)
}

func (a Assignment) GetDepth() int {
	if a.Parent == nil {
		return 0
	} else {
		return a.Parent.GetDepth() + 1
	}
}

func (a Assignment) MaxDepth() (d int) {
	d = a.GetDepth()

	for _, c := range a.Children {
		cd := c.MaxDepth()

		if d < cd {
			d = cd
		}
	}

	return d
}

func (a Assignment) AlignmentSpacing() int {
	if a.Transacted.Metadata.GetTags().Len() == 1 && ids.IsDependentLeaf(a.Transacted.Metadata.GetTags().Any()) {
		return a.Parent.AlignmentSpacing() + len(
			a.Parent.Transacted.Metadata.GetTags().Any().String(),
		)
	}

	return 0
}

func (a Assignment) MaxLen() (m int) {
	for _, z := range a.Objects.All() {
		oM := z.sku.GetSkuExternal().ObjectId.Len()

		if oM > m {
			m = oM
		}
	}

	for _, c := range a.Children {
		oM := c.MaxLen()

		if oM > m {
			m = oM
		}
	}

	return m
}

func (a Assignment) String() (s string) {
	if a.Parent != nil {
		s = a.Parent.String() + "."
	}

	return s + quiter.StringCommaSeparated(a.Transacted.Metadata.GetTags())
}

func (a *Assignment) makeChild(e ids.Tag) (b *Assignment) {
	b = newAssignment(a.GetDepth() + 1)
	b.Transacted.GetMetadataMutable().SetTags(ids.MakeMutableTagSet(e))
	a.addChild(b)
	return b
}

func (a *Assignment) makeChildWithSet(es ids.TagMutableSet) (b *Assignment) {
	b = newAssignment(a.GetDepth() + 1)
	b.Transacted.GetMetadataMutable().SetTags(es)
	a.addChild(b)
	return b
}

func (a *Assignment) addChild(c *Assignment) {
	if a == c {
		panic("child and parent are the same")
	}

	if c.Parent != nil && c.Parent == a {
		panic("child already has self as parent")
	}

	if c.Parent != nil {
		panic("child already has a parent")
	}

	a.Children = append(a.Children, c)
	c.Parent = a
}

func (a *Assignment) parentOrRoot() (p *Assignment) {
	switch a.Parent {
	case nil:
		return a

	default:
		return a.Parent
	}
}

func (a *Assignment) nthParent(n int) (p *Assignment, err error) {
	if n < 0 {
		n = -n
	}

	if n == 0 {
		p = a
		return p, err
	}

	if a.Parent == nil {
		err = errors.ErrorWithStackf("cannot get nth parent as parent is nil")
		return p, err
	}

	return a.Parent.nthParent(n - 1)
}

func (a *Assignment) removeFromParent() (err error) {
	return a.Parent.removeChild(a)
}

func (a *Assignment) removeChild(c *Assignment) (err error) {
	if c.Parent != a {
		err = errors.ErrorWithStackf("attempting to remove child from wrong parent")
		return err
	}

	if len(a.Children) == 0 {
		err = errors.ErrorWithStackf(
			"attempting to remove child when there are no children",
		)
		return err
	}

	cap1 := 0
	cap2 := len(a.Children) - 1

	if cap2 > 0 {
		cap1 = cap2
	}

	nc := make([]*Assignment, 0, cap1)

	for _, c1 := range a.Children {
		if c1 == c {
			continue
		}

		nc = append(nc, c1)
	}

	c.Parent = nil
	a.Children = nc

	return err
}

func (a *Assignment) consume(b *Assignment) (err error) {
	for _, c := range b.Children {
		if err = c.removeFromParent(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		a.addChild(c)
	}

	for _, obj := range b.Objects.All() {
		a.AddObject(obj)
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

func (a *Assignment) AllTags(mes ids.TagMutableSet) (err error) {
	if a == nil {
		return err
	}

	var es ids.TagSet

	if es, err = a.expandedTags(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for e := range es.AllPtr() {
		if err = mes.AddPtr(e); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if err = a.Parent.AllTags(mes); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (a *Assignment) expandedTags() (es ids.TagSet, err error) {
	es = ids.MakeTagSet()

	if a.Transacted.GetMetadata().GetTags().Len() != 1 || a.Parent == nil {
		es = a.Transacted.GetMetadata().GetTags().CloneSetPtrLike()
		return es, err
	} else {
		e := a.Transacted.GetMetadata().GetTags().Any()

		if ids.IsDependentLeaf(e) {
			var pe ids.TagSet

			if pe, err = a.Parent.expandedTags(); err != nil {
				err = errors.Wrap(err)
				return es, err
			}

			if pe.Len() > 1 {
				err = errors.ErrorWithStackf(
					"cannot infer full tag for assignment because parent assignment has more than one tags: %s",
					a.Parent.Transacted.GetMetadata().GetTags(),
				)

				return es, err
			}

			e1 := pe.Any()

			if ids.IsEmpty(e1) {
				err = errors.ErrorWithStackf("parent tag is empty")
				return es, err
			}

			if err = e.Set(fmt.Sprintf("%s%s", e1, e)); err != nil {
				err = errors.Wrap(err)
				return es, err
			}
		}

		es = ids.MakeTagSet(e)
	}

	return es, err
}

func (a *Assignment) SubtractFromSet(es ids.TagMutableSet) (err error) {
	for e := range a.Transacted.GetMetadata().GetTags().AllPtr() {
		for e1 := range es.AllPtr() {
			if ids.ContainsExactly(e1, e) {
				if err = es.DelPtr(e1); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}
		}

		if err = es.DelPtr(e); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if a.Parent == nil {
		return err
	}

	return a.Parent.SubtractFromSet(es)
}

func (a *Assignment) Contains(e *ids.Tag) bool {
	if a.Transacted.GetMetadata().GetTags().ContainsKey(e.String()) {
		return true
	}

	if a.Parent == nil {
		return false
	}

	return a.Parent.Contains(e)
}

func (parent *Assignment) SortChildren() {
	sort.Slice(parent.Children, func(i, j int) bool {
		esi := parent.Children[i].Transacted.GetMetadata().GetTags()
		esj := parent.Children[j].Transacted.GetMetadata().GetTags()

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
