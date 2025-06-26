package store_config

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func init() {
	collections_value.RegisterGobValue[*tag](nil)
}

type implicitTagMap map[string]ids.TagMutableSet

func (iem implicitTagMap) Contains(to, imp ids.Tag) bool {
	s, ok := iem[to.String()]

	if !ok || s == nil {
		return false
	}

	if !s.Contains(imp) {
		return false
	}

	return true
}

func (iem implicitTagMap) Set(to, imp ids.Tag) (err error) {
	s, ok := iem[to.String()]

	if !ok {
		s = ids.MakeTagMutableSet()
		iem[to.String()] = s
	}

	return s.Add(imp)
}

type tag struct {
	Transacted sku.Transacted
	Computed   bool
}

func (a *tag) Less(b *tag) bool {
	return sku.TransactedLessor.Less(&a.Transacted, &b.Transacted)
}

func (a *tag) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a *tag) Equals(b *tag) bool {
	if !a.Transacted.Equals(&b.Transacted) {
		return false
	}

	if !quiter.SetEqualsPtr(
		a.Transacted.Metadata.Cache.GetImplicitTags(),
		b.Transacted.Metadata.Cache.GetImplicitTags(),
	) {
		return false
	}

	return true
}

func (e *tag) Set(v string) (err error) {
	if err = e.Transacted.ObjectId.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (e *tag) String() string {
	return e.Transacted.GetObjectId().String()
}

func (k *compiled) AccumulateImplicitTags(
	e ids.Tag,
) (err error) {
	ek, ok := k.Tags.Get(e.String())

	if !ok {
		return
	}

	ees := ids.MakeTagMutableSet()

	ids.ExpandOneInto(
		e,
		ids.MakeTag,
		expansion.ExpanderRight,
		ees,
	)

	if err = ees.Each(
		func(e1 ids.Tag) (err error) {
			if e1.Equals(e) {
				return
			}

			if err = k.AccumulateImplicitTags(e1); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = k.GetImplicitTags(&e1).Each(
				func(e2 ids.Tag) (err error) {
					return k.ImplicitTags.Set(e, e2)
				},
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = ek.Transacted.Metadata.GetTags().Each(
		func(e1 ids.Tag) (err error) {
			if k.ImplicitTags.Contains(e1, e) {
				return
			}

			if err = k.ImplicitTags.Set(e, e1); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = k.AccumulateImplicitTags(e1); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (k *compiled) addTag(
	kinder *sku.Transacted,
	mutter *sku.Transacted,
) (didChange bool, err error) {
	k.lock.Lock()
	defer k.lock.Unlock()

	var b tag

	sku.Resetter.ResetWith(&b.Transacted, kinder)

	if didChange, err = quiter.AddOrReplaceIfGreater(k.Tags, &b); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
