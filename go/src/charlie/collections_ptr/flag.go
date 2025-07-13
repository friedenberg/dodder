package collections_ptr

import (
	"flag"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

type SetterPolicy int

const (
	SetterPolicyAppend = SetterPolicy(iota)
	SetterPolicyReset
)

type flagPtr[T interfaces.ValueLike] interface {
	interfaces.ValuePtr[T]
	interfaces.SetterPtr[T]
}

// TODO-P2 add Resetter2 and Pool
type Flag[
	T interfaces.ValueLike,
	TPtr flagPtr[T],
] interface {
	flag.Value
	SetMany(vs ...string) (err error)
	interfaces.MutableSetPtrLike[T, TPtr]
	GetSetPtrLike() interfaces.SetPtrLike[T, TPtr]
	GetMutableSetPtrLike() interfaces.MutableSetPtrLike[T, TPtr]
}

func MakeFlagCommasFromExisting[
	T interfaces.ValueLike,
	TPtr flagPtr[T],
](
	p SetterPolicy,
	existing interfaces.MutableSetPtrLike[T, TPtr],
	// pool schnittstellen.Pool[T, TPtr],
	// resetter schnittstellen.Resetter2[T, TPtr],
) Flag[T, TPtr] {
	return &flagCommas[T, TPtr]{
		SP:                p,
		MutableSetPtrLike: existing,
		// pool:              pool,
		// resetter:          resetter,
	}
}

func MakeFlagCommas[
	T interfaces.ValueLike,
	TPtr flagPtr[T],
](
	p SetterPolicy,
	// pool schnittstellen.Pool[T, TPtr],
	// resetter schnittstellen.Resetter2[T, TPtr],
) Flag[T, TPtr] {
	return &flagCommas[T, TPtr]{
		SP:                p,
		MutableSetPtrLike: MakeMutableValueSet[T, TPtr](nil),
		// pool:              pool,
		// resetter:          resetter,
	}
}

type flagCommas[
	T interfaces.ValueLike,
	TPtr flagPtr[T],
] struct {
	SP SetterPolicy
	interfaces.MutableSetPtrLike[T, TPtr]
	pool     interfaces.Pool[T, TPtr]
	resetter interfaces.Resetter2[T, TPtr]
}

func (f flagCommas[T, TPtr]) All() interfaces.Seq[T] {
	return f.MutableSetPtrLike.All()
}

func (f flagCommas[T, TPtr]) GetSetPtrLike() (s interfaces.SetPtrLike[T, TPtr]) {
	return f.CloneSetPtrLike()
}

func (f flagCommas[T, TPtr]) GetMutableSetPtrLike() (s interfaces.MutableSetPtrLike[T, TPtr]) {
	return f.CloneMutableSetPtrLike()
}

func (f flagCommas[T, TPtr]) String() (out string) {
	if f.MutableSetPtrLike == nil {
		return
	}

	sorted := quiter.SortedStrings[T](f)

	sb := &strings.Builder{}
	first := true

	for _, e1 := range sorted {
		if !first {
			sb.WriteString(", ")
		}

		sb.WriteString(e1)

		first = false
	}

	out = sb.String()

	return
}

func (f *flagCommas[T, TPtr]) SetMany(vs ...string) (err error) {
	for _, v := range vs {
		if err = f.Set(v); err != nil {
			return
		}
	}

	return
}

func (f *flagCommas[T, TPtr]) Set(v string) (err error) {
	switch f.SP {
	case SetterPolicyReset:
		f.Reset()
	}

	els := strings.Split(v, ",")

	for _, e := range els {
		e = strings.TrimSpace(e)

		// TODO-P2 use iter.AddStringPtr
		if err = quiter.AddString[T, TPtr](f, e); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
