package collections_ptr

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

type SetterPolicy int

const (
	SetterPolicyAppend = SetterPolicy(iota)
	SetterPolicyReset
)

// TODO-P2 add Resetter2 and Pool
type Flag[
	VALUE interfaces.Value[VALUE],
	VALUE_PTR interfaces.ValuePtr[VALUE],
] interface {
	interfaces.FlagValue
	SetMany(vs ...string) (err error)
	interfaces.MutableSetPtrLike[VALUE, VALUE_PTR]
	GetSetPtrLike() interfaces.SetPtrLike[VALUE, VALUE_PTR]
	GetMutableSetPtrLike() interfaces.MutableSetPtrLike[VALUE, VALUE_PTR]
}

func MakeFlagCommasFromExisting[
	VALUE interfaces.Value[VALUE],
	VALUE_PTR interfaces.ValuePtr[VALUE],
](
	policy SetterPolicy,
	existing interfaces.MutableSetPtrLike[VALUE, VALUE_PTR],
) Flag[VALUE, VALUE_PTR] {
	return &flagCommas[VALUE, VALUE_PTR]{
		SetterPolicy:      policy,
		MutableSetPtrLike: existing,
	}
}

func MakeFlagCommas[
	VALUE interfaces.Value[VALUE],
	VALUE_PTR interfaces.ValuePtr[VALUE],
](
	policy SetterPolicy,
) Flag[VALUE, VALUE_PTR] {
	return &flagCommas[VALUE, VALUE_PTR]{
		SetterPolicy:      policy,
		MutableSetPtrLike: MakeMutableValueSet[VALUE, VALUE_PTR](nil),
	}
}

type flagCommas[
	VALUE interfaces.Value[VALUE],
	VALUE_PTR interfaces.ValuePtr[VALUE],
] struct {
	SetterPolicy SetterPolicy
	interfaces.MutableSetPtrLike[VALUE, VALUE_PTR]
	pool     interfaces.Pool[VALUE, VALUE_PTR]
	resetter interfaces.ResetterPtr[VALUE, VALUE_PTR]
}

func (flags flagCommas[T, TPtr]) All() interfaces.Seq[T] {
	return flags.MutableSetPtrLike.All()
}

func (flags flagCommas[T, TPtr]) GetSetPtrLike() (s interfaces.SetPtrLike[T, TPtr]) {
	return flags.CloneSetPtrLike()
}

func (flags flagCommas[T, TPtr]) GetMutableSetPtrLike() (s interfaces.MutableSetPtrLike[T, TPtr]) {
	return flags.CloneMutableSetPtrLike()
}

func (flags flagCommas[T, TPtr]) String() (out string) {
	if flags.MutableSetPtrLike == nil {
		return out
	}

	sorted := quiter.SortedStrings[T](flags)

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

	return out
}

func (flags *flagCommas[T, TPtr]) SetMany(vs ...string) (err error) {
	for _, v := range vs {
		if err = flags.Set(v); err != nil {
			return err
		}
	}

	return err
}

func (flags *flagCommas[T, TPtr]) Set(v string) (err error) {
	switch flags.SetterPolicy {
	case SetterPolicyReset:
		flags.Reset()
	}

	els := strings.Split(v, ",")

	for _, e := range els {
		e = strings.TrimSpace(e)

		// TODO-P2 use iter.AddStringPtr
		if err = quiter.AddString[T, TPtr](flags, e); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
