package quiter

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MakeFuncSetString[
	E any,
	EPtr interfaces.SetterPtr[E],
](
	c interfaces.Adder[E],
) interfaces.FuncSetString {
	return func(v string) (err error) {
		return AddString[E, EPtr](c, v)
	}
}

func Len(cs ...interfaces.Lenner) (n int) {
	for _, c := range cs {
		n += c.Len()
	}

	return
}

func DerivedValues[E any, F any](
	c interfaces.SetLike[E],
	f interfaces.FuncTransform[E, F],
) (out []F, err error) {
	out = make([]F, 0, c.Len())

	for e := range c.All() {
		var e1 F

		if e1, err = f(e); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return
		}

		out = append(out, e1)
	}

	return
}
