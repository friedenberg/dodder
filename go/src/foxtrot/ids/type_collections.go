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
	TypeSet        = interfaces.SetPtrLike[Type, *Type]
	TypeMutableSet = interfaces.MutableSetPtrLike[Type, *Type]
)

func MakeTypSet(es ...Type) (s TypeSet) {
	return TypeSet(
		collections_ptr.MakeValueSetValue[Type, *Type](nil, es...),
	)
}

func MakeTypSetStrings(vs ...string) (s TypeSet, err error) {
	return collections_ptr.MakeValueSetString[Type, *Type](nil, vs...)
}

func MakeMutableTypeSet(hs ...Type) TypeMutableSet {
	return MakeTypMutableSet(hs...)
}

func MakeTypMutableSet(hs ...Type) TypeMutableSet {
	return TypeMutableSet(
		collections_ptr.MakeMutableValueSetValue[Type, *Type](
			nil,
			hs...,
		),
	)
}
