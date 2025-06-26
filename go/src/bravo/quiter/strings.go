package quiter

import (
	"sort"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func SortedValuesBy[E any](
	c interfaces.SetLike[E],
	sf func(E, E) bool,
) (out []E) {
	out = Elements(c)

	sort.Slice(out, func(i, j int) bool { return sf(out[i], out[j]) })

	return
}

func SortedValues[E interfaces.Value[E]](
	c interfaces.SetLike[E],
) (out []E) {
	out = Elements(c)

	sort.Slice(
		out,
		func(i, j int) bool { return out[i].String() < out[j].String() },
	)

	return
}

func Strings[E interfaces.Stringer](
	cs ...interfaces.Collection[E],
) (out []string) {
	l := 0

	for _, c := range cs {
		if c == nil {
			continue
		}

		l += c.Len()
	}

	out = make([]string, 0, l)

	for _, c := range cs {
		if c == nil {
			continue
		}

		for e := range c.All() {
			out = append(out, e.String())
		}
	}

	return
}

func SortedStrings[E interfaces.Stringer](
	cs ...interfaces.Collection[E],
) (out []string) {
	out = Strings(cs...)

	sort.Strings(out)

	return
}

func StringDelimiterSeparated[E interfaces.Stringer](
	d string,
	cs ...interfaces.Collection[E],
) string {
	if cs == nil {
		return ""
	}

	sorted := SortedStrings[E](cs...)

	if len(sorted) == 0 {
		return ""
	}

	sb := &strings.Builder{}
	first := true

	for _, e1 := range sorted {
		if !first {
			sb.WriteString(d)
		}

		sb.WriteString(e1)

		first = false
	}

	return sb.String()
}

func StringCommaSeparated[E interfaces.Stringer](
	cs ...interfaces.Collection[E],
) string {
	return StringDelimiterSeparated(", ", cs...)
}

func ReverseSortable(s sort.Interface) {
	max := s.Len() / 2

	for i := 0; i < max; i++ {
		s.Swap(i, s.Len()-1-i)
	}
}
