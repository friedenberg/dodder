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

func AddOrReplaceIfGreater[T interface {
	interfaces.Stringer
	interfaces.ValueLike
	interfaces.Lessable[T]
}](
	c interfaces.SetMutable[T],
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

// Constructs an object of type `ELEMENT` by using its `Set` method and adds it to
// the given `adder`
func AddString[ELEMENT any, ELEMENT_PTR interfaces.SetterPtr[ELEMENT]](
	adder interfaces.Adder[ELEMENT],
	value string,
) (err error) {
	var element ELEMENT

	if err = ELEMENT_PTR(&element).Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = adder.Add(element); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
