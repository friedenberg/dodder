package quiter

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func AppendSeq[ELEMENT any, APPENDER interface{ Append(...ELEMENT) }](
	collection APPENDER,
	seq interfaces.Seq[ELEMENT],
) {
	for element := range seq {
		collection.Append(element)
	}
}

func AppendSeq2[INDEX any, ELEMENT any, APPENDER interface{ Append(...ELEMENT) }](
	collection APPENDER,
	seq interfaces.Seq2[INDEX, ELEMENT],
) {
	for _, element := range seq {
		collection.Append(element)
	}
}

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
	r interfaces.ResetterPtr[E, EPtr],
	b EPtr,
) (err error) {
	a := p.Get()
	r.ResetWith(a, b)
	return s.AddPtr(a)
}

func AddOrReplaceIfGreater[T interface {
	interfaces.Stringer
	interfaces.ValueLike
	interfaces.Lessable[T]
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
