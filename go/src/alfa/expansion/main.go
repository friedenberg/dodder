package expansion

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

var (
	ExpanderRight = MakeExpanderRight(`-`)
	ExpanderAll   = MakeExpanderAll(`-`)
)

type Expander interface {
	Expand(string) interfaces.Seq[string]
}
