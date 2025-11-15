package flag

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

func MakeResettingFlag[T interface {
	interfaces.Stringer
},
	TPtr interface {
		interfaces.SetterPtr[T]
		Reset()
	},
](e T) ResettingFlag[T, TPtr] {
	return ResettingFlag[T, TPtr]{
		flag: e,
	}
}

type ResettingFlag[T interface {
	interfaces.Stringer
},
	TPtr interface {
		interfaces.SetterPtr[T]
		Reset()
	},
] struct {
	wasSet bool
	flag   T
}

func (g *ResettingFlag[T, TPtr]) Set(v string) (err error) {
	if !g.wasSet {
		TPtr(&g.flag).Reset()
	}

	g.wasSet = true

	return TPtr(&g.flag).Set(v)
}

func (g *ResettingFlag[T, TPtr]) String() string {
	return g.flag.String()
}

func (g ResettingFlag[T, TPtr]) GetFlag() T {
	return g.flag
}

func (g *ResettingFlag[T, TPtr]) GetFlagPtr() TPtr {
	return &g.flag
}
