package expansion

import (
	"regexp"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
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
