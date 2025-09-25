package quiter

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func AddClone[E any, EPtr interface {
	*E
	ResetWithPtr(*E)
}](
	c interfaces.Adder[EPtr],
) interfaces.FuncIter[EPtr] {
	return func(e EPtr) (err error) {
		var e1 E
		EPtr(&e1).ResetWithPtr((*E)(e))
		c.Add(&e1)
		return err
	}
}

func AddClonePool[E any, EPtr interfaces.Ptr[E]](
	s interfaces.AdderPtr[E, EPtr],
	p interfaces.Pool[E, EPtr],
	r interfaces.Resetter2[E, EPtr],
	b EPtr,
) (err error) {
	a := p.Get()
	r.ResetWith(a, b)
	return s.AddPtr(a)
}

func AddOrReplaceIfGreater[T interface {
	interfaces.Stringer
	interfaces.ValueLike
	interfaces.Lessor[T]
}](
	c interfaces.MutableSetLike[T],
	b T,
) (shouldAdd bool, err error) {
	a, ok := c.Get(c.Key(b))

	// 	if ok {
	// 		log.Debug().Print("less:", a.Less(b))
	// 	} else {
	// 		log.Debug().Print("ok:", ok)
	// 	}

	shouldAdd = !ok || a.Less(b)

	if shouldAdd {
		err = c.Add(b)
	}

	return shouldAdd, err
}

func AddString[E any, EPtr interfaces.SetterPtr[E]](
	c interfaces.Adder[E],
	v string,
) (err error) {
	var e E

	if err = EPtr(&e).Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = c.Add(e); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
