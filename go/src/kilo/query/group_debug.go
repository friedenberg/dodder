package query

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

func (qg *Query) StringDebug() string {
	var sb strings.Builder

	if qg.defaultQuery != nil {
		fmt.Fprintf(&sb, "default: %q", qg.defaultQuery)
	}

	first := true

	for _, g := range qg.sortedUserQueries() {
		if !first {
			sb.WriteRune(' ')
		}

		sb.WriteString(g.StringDebug())

		first = false
	}

	sb.WriteString(" | ")
	first = true

	for _, g := range genres.All() {
		q, ok := qg.optimizedQueries[g]

		if !ok {
			continue
		}

		if !first {
			sb.WriteRune(' ')
		}

		sb.WriteString(q.String())

		first = false
	}

	return sb.String()
}

func (qg *Query) StringOptimized() string {
	var sb strings.Builder

	first := true

	// qg.FDs.Each(
	// 	func(f *fd.FD) error {
	// 		if !first {
	// 			sb.WriteRune(' ')
	// 		}

	// 		sb.WriteString(f.String())

	// 		first = false

	// 		return nil
	// 	},
	// )

	for _, g := range genres.All() {
		q, ok := qg.optimizedQueries[g]

		if !ok {
			continue
		}

		if !first {
			sb.WriteRune(' ')
		}

		sb.WriteString(q.String())

		first = false
	}

	return sb.String()
}
