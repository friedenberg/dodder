package ids

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
)

func init() {
	collections_value.RegisterGobValue[Tag](nil)
}

type (
	ZettelIdSet        = interfaces.SetLike[ZettelId]
	ZettelIdMutableSet = interfaces.MutableSetLike[ZettelId]
)

func MakeZettelIdMutableSet(hs ...ZettelId) ZettelIdMutableSet {
	return ZettelIdMutableSet(
		collections_value.MakeMutableValueSet(nil, hs...),
	)
}
