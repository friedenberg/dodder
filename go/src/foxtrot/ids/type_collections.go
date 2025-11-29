package ids

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/thyme"
)

func init() {
	collections_value.RegisterGobValue[thyme.Time](nil)
}

type (
	TypeSet        = interfaces.Set[Type]
	TypeMutableSet = interfaces.SetMutable[Type]
)

func MakeTypeSet(es ...Type) (s TypeSet) {
	return TypeSet(collections_ptr.MakeValueSetValue(nil, es...))
}

func MakeTypeSetStrings(vs ...string) (s TypeSet, err error) {
	return collections_ptr.MakeValueSetString[Type](nil, vs...)
}

func MakeMutableTypeSet(hs ...Type) TypeMutableSet {
	return MakeTypeMutableSet(hs...)
}

func MakeTypeMutableSet(hs ...Type) TypeMutableSet {
	return TypeMutableSet(collections_ptr.MakeMutableValueSetValue(nil, hs...))
}
