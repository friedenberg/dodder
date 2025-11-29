package ids

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
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

	for _, e := range quiter.CollectSlice[Tag](haystack) {
		if strings.HasPrefix(e.String(), needle.String()) {
			s4.Add(e)
		}
	}

	s3 = CloneTagSet(s4)

	return s3
}

func SubtractPrefix(s1 TagSet, e Tag) (s2 TagSet) {
	s3 := MakeTagMutableSet()

	for _, e1 := range quiter.CollectSlice[Tag](s1) {
		e2, _ := LeftSubtract(e1, e)

		if e2.String() == "" {
			continue
		}

		s3.Add(e2)
	}

	s2 = CloneTagSet(s3)

	return s2
}

func WithRemovedCommonPrefixes(tags TagSet) (output TagSet) {
	sortedTags := quiter.SortedValues[Tag](tags.All())
	filteredTags := make([]Tag, 0, len(sortedTags))

	for _, e := range sortedTags {
		if len(filteredTags) == 0 {
			filteredTags = append(filteredTags, e)
			continue
		}

		idxLast := len(filteredTags) - 1
		last := filteredTags[idxLast]

		switch {
		case Contains(last, e):
			continue

		case Contains(e, last):
			filteredTags[idxLast] = e

		default:
			filteredTags = append(filteredTags, e)
		}
	}

	output = MakeTagSetFromSlice(filteredTags...)

	return output
}

func AddNormalizedTag(es TagSetMutable, e *Tag) {
	ExpandOneInto(
		*e,
		MakeTag,
		expansion.ExpanderRight,
		es,
	)

	c := CloneTagSet(es)
	es.Reset()
	for tag := range WithRemovedCommonPrefixes(c).All() {
		es.Add(tag)
	}
}

func RemovePrefixes(haystack TagSetMutable, needle Tag) {
	for tag := range haystack.All() {
		if !strings.HasPrefix(tag.String(), needle.String()) {
			continue
		}

		quiter_set.Del(haystack, tag)
	}
}
