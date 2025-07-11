package values

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func Equals[T interfaces.Equatable[T]](a T, b any) bool {
	{
		b1, ok := b.(T)

		if ok {
			return a.Equals(b1)
		}
	}

	return false
}
