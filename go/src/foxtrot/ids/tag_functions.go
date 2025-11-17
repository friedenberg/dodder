package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

func IsDependentLeaf(a Tag) (has bool) {
	has = strings.HasPrefix(strings.TrimSpace(a.String()), "-")
	return has
}

func HasParentPrefix(a, b Tag) (has bool) {
	has = strings.HasPrefix(strings.TrimSpace(a.String()), b.String())
	return has
}

func IntersectPrefixes(haystack TagSet, needle Tag) (s3 TagSet) {
	s4 := MakeTagMutableSet()

	for _, e := range quiter.Elements[Tag](haystack) {
		if strings.HasPrefix(e.String(), needle.String()) {
			s4.Add(e)
		}
	}

	s3 = s4.CloneSetPtrLike()

	return s3
}

func SubtractPrefix(s1 TagSet, e Tag) (s2 TagSet) {
	s3 := MakeTagMutableSet()

	for _, e1 := range quiter.Elements[Tag](s1) {
		e2, _ := LeftSubtract(e1, e)

		if e2.String() == "" {
			continue
		}

		s3.Add(e2)
	}

	s2 = s3.CloneSetPtrLike()

	return s2
}

func WithRemovedCommonPrefixes(s TagSet) (s2 TagSet) {
	es1 := quiter.SortedValues[Tag](s)
	es := make([]Tag, 0, len(es1))

	for _, e := range es1 {
		if len(es) == 0 {
			es = append(es, e)
			continue
		}

		idxLast := len(es) - 1
		last := es[idxLast]

		switch {
		case Contains(last, e):
			continue

		case Contains(e, last):
			es[idxLast] = e

		default:
			es = append(es, e)
		}
	}

	s2 = MakeTagSet(es...)

	return s2
}

func AddNormalizedTag(es TagMutableSet, e *Tag) {
	ExpandOneInto(
		*e,
		MakeTag,
		expansion.ExpanderRight,
		es,
	)

	c := es.CloneSetPtrLike()
	es.Reset()
	for tag := range WithRemovedCommonPrefixes(c).All() {
		es.Add(tag)
	}
}

func RemovePrefixes(es TagMutableSet, needle Tag) {
	for _, haystack := range quiter.Elements(es) {
		// TODO-P2 make more efficient
		if strings.HasPrefix(haystack.String(), needle.String()) {
			es.Del(haystack)
		}
	}
}
