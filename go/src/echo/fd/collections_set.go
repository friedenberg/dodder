package fd

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
)

func init() {
	collections_value.RegisterGobValue[*FD](nil)
}

type (
	Set        = interfaces.SetLike[*FD]
	MutableSet = interfaces.MutableSetLike[*FD]
)

func MakeSet(ts ...*FD) Set {
	return collections_value.MakeValueSet[*FD](
		nil,
		ts...,
	)
}

func MakeMutableSet(ts ...*FD) MutableSet {
	return collections_value.MakeMutableValueSet[*FD](
		nil,
		ts...,
	)
}

func MakeMutableSetSha() MutableSet {
	return collections_value.MakeMutableValueSet[*FD](
		KeyerSha{},
	)
}

type KeyerSha struct{}

func (KeyerSha) GetKey(fd *FD) string {
	return fd.sha.String()
}
