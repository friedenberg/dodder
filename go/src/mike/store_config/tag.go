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
		return err
	}

	return err
}

func (e *tag) String() string {
	return e.Transacted.GetObjectId().String()
}

func (compiled *compiled) AccumulateImplicitTags(
	e ids.Tag,
) (err error) {
	ek, ok := compiled.Tags.Get(e.String())

	if !ok {
		return err
	}

	ees := ids.MakeTagMutableSet()

	ids.ExpandOneInto(
		e,
		ids.MakeTag,
		expansion.ExpanderRight,
		ees,
	)

	for e1 := range ees.All() {
		if e1.Equals(e) {
			continue
		}

		if err = compiled.AccumulateImplicitTags(e1); err != nil {
			err = errors.Wrap(err)
			return err
		}

		for e2 := range compiled.GetImplicitTags(&e1).All() {
			if err = compiled.ImplicitTags.Set(e, e2); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	for e1 := range ek.Transacted.Metadata.GetTags().All() {
		if compiled.ImplicitTags.Contains(e1, e) {
			continue
		}

		if err = compiled.ImplicitTags.Set(e, e1); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = compiled.AccumulateImplicitTags(e1); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (compiled *compiled) addTag(
	kinder *sku.Transacted,
	mutter *sku.Transacted,
) (didChange bool, err error) {
	compiled.lock.Lock()
	defer compiled.lock.Unlock()

	var b tag

	sku.Resetter.ResetWith(&b.Transacted, kinder)

	if didChange, err = quiter.AddOrReplaceIfGreater(compiled.Tags, &b); err != nil {
		err = errors.Wrap(err)
		return didChange, err
	}

	return didChange, err
}
