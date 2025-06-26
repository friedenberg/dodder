package expansion

import (
	"regexp"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type expanderRight struct {
	delimiter *regexp.Regexp
}

func MakeExpanderRight(
	delimiter string,
) expanderRight {
	return expanderRight{
		delimiter: regexp.MustCompile(delimiter),
	}
}

func (ex expanderRight) Expand(
	sa interfaces.FuncSetString,
	s string,
) {
	sa(s)

	if s == "" {
		return
	}

	delim := ex.delimiter.FindAllIndex([]byte(s), -1)

	if delim == nil {
		return
	}

	for _, loc := range delim {
		locStart := loc[0]
		t1 := s[0:locStart]

		sa(t1)
	}
}
