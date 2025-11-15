package expansion

import (
	"regexp"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

var (
	regexExpandTagsHyphens *regexp.Regexp
	ExpanderRight          Expander
	ExpanderAll            Expander
)

type Expander interface {
	Expand(interfaces.FuncSetString, string)
}

func init() {
	regexExpandTagsHyphens = regexp.MustCompile(`-`)
	ExpanderRight = MakeExpanderRight(`-`)
	ExpanderAll = MakeExpanderAll(`-`)
}
