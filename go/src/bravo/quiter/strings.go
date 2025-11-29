package quiter

import (
	"slices"
	"sort"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

func SortedValuesBy[ELEMENT any](
	set interfaces.SetLike[ELEMENT],
	sortFunc func(ELEMENT, ELEMENT) bool,
) (out []ELEMENT) {
	out = CollectSlice(set)

	sort.Slice(out, func(i, j int) bool { return sortFunc(out[i], out[j]) })

	return out
}

func SortedValues[ELEMENT interfaces.Value[ELEMENT]](
	seq interfaces.Seq[ELEMENT],
) (out []ELEMENT) {
	out = slices.Collect(seq)

	sort.Slice(
		out,
		func(i, j int) bool { return out[i].String() < out[j].String() },
	)

	return out
}

func Strings[ELEMENT interfaces.Stringer](
	collections ...interfaces.Collection[ELEMENT],
) (out []string) {
	l := 0

	for _, c := range collections {
		if c == nil {
			continue
		}

		l += c.Len()
	}

	out = make([]string, 0, l)

	for _, c := range collections {
		if c == nil {
			continue
		}

		for e := range c.All() {
			out = append(out, e.String())
		}
	}

	return out
}

func SortedStrings[ELEMENT interfaces.Stringer](
	collections ...interfaces.Collection[ELEMENT],
) (out []string) {
	out = Strings(collections...)

	sort.Strings(out)

	return out
}

func StringDelimiterSeparated[ELEMENT interfaces.Stringer](
	delimiter string,
	collections ...interfaces.Collection[ELEMENT],
) string {
	if collections == nil {
		return ""
	}

	sorted := SortedStrings[ELEMENT](collections...)

	if len(sorted) == 0 {
		return ""
	}

	sb := &strings.Builder{}
	first := true

	for _, e1 := range sorted {
		if !first {
			sb.WriteString(delimiter)
		}

		sb.WriteString(e1)

		first = false
	}

	return sb.String()
}

func StringCommaSeparated[ELEMENT interfaces.Stringer](
	collections ...interfaces.Collection[ELEMENT],
) string {
	return StringDelimiterSeparated(", ", collections...)
}

func ReverseSortable(sortable sort.Interface) {
	max := sortable.Len() / 2

	for i := range max {
		sortable.Swap(i, sortable.Len()-1-i)
	}
}
