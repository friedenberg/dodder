package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/object_id_provider"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// Saves the blob if necessary, applies the proto object, runs pre-commit hooks,
// runs the new hook, validates the blob, then calculates the sha for the object
func (s *Store) tryPrecommit(
	external sku.ExternalLike,
	parent *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if err = s.SaveBlob(external); err != nil {
		err = errors.Wrap(err)
		return
	}

	kinder := external.GetSku()

	if parent == nil {
		options.Proto.Apply(kinder, kinder)
	}

	// TODO decide if the type proto should actually be applied every time
	if options.ApplyProtoType {
		s.protoZettel.ApplyType(kinder, kinder)
	}

	if genres.Type == external.GetSku().GetGenre() {
		if external.GetSku().GetType().IsEmpty() {
			external.GetSku().GetMetadata().Type = ids.DefaultOrPanic(
				genres.Type,
			)
		}
	}

	// modify pre commit hooks to support import
	if err = s.tryPreCommitHooks(kinder, parent, options); err != nil {
		if s.config.GetCLIConfig().IgnoreHookErrors {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	// TODO just just mutter == nil
	if parent == nil {
		if err = s.tryNewHook(kinder, options); err != nil {
			if s.config.GetCLIConfig().IgnoreHookErrors {
				err = nil
			} else {
				err = errors.Wrap(err)
				return
			}
		}
	}

	if err = s.validate(kinder, parent, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = kinder.CalculateObjectShas(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO add RealizeAndOrStore result
func (store *Store) Commit(
	external sku.ExternalLike,
	options sku.CommitOptions,
) (err error) {
	child := external.GetSku()

	ui.Log().Printf("%s -> %s", options, child)

	if !store.GetEnvRepo().GetLockSmith().IsAcquired() &&
		(options.AddToInventoryList || options.StreamIndexOptions.AddToStreamIndex) {
		err = errors.Wrap(file_lock.ErrLockRequired{
			Operation: "commit",
		})

		return
	}

	// TAI must be set before calculating object sha
	if options.UpdateTai {
		if options.Clock == nil {
			options.Clock = store
		}

		child.SetTai(options.Clock.GetTai())
	}

	if options.AddToInventoryList && (child.ObjectId.IsEmpty() ||
		child.GetGenre() == genres.None ||
		child.GetGenre() == genres.Blob) {
		var zettelId *ids.ZettelId

		if zettelId, err = store.zettelIdIndex.CreateZettelId(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = child.ObjectId.SetWithIdLike(zettelId); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var parent *sku.Transacted

	if parent, err = store.fetchParentIfNecessary(child); err != nil {
		err = errors.Wrap(err)
		return
	}

	if parent != nil {
		defer sku.GetTransactedPool().Put(parent)
		child.Metadata.Cache.ParentTai = parent.GetTai()
	}

	if err = store.tryPrecommit(external, parent, options); err != nil {
		err = errors.Wrap(err)
		return
	}

	if options.AddToInventoryList {
		if err = store.addMissingTypeAndTags(options, child); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if options.AddToInventoryList ||
		options.StreamIndexOptions.AddToStreamIndex {
		if err = store.addObjectToAbbrStore(child); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// short circuits if the parent is equal to the child
	if options.AddToInventoryList &&
		parent != nil &&
		ids.Equals(child.GetObjectId(), parent.GetObjectId()) &&
		child.Metadata.EqualsSansTai(&parent.Metadata) {

		sku.TransactedResetter.ResetWithExceptFields(child, parent)

		if store.sunrise.Less(child.GetTai()) {
			if err = store.ui.TransactedUnchanged(child); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		return
	}

	if err = store.applyDormantAndRealizeTags(
		child,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if options.AddToInventoryList ||
		options.StreamIndexOptions.AddToStreamIndex {
		if err = store.config.AddTransacted(
			child,
			parent,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if child.GetGenre() == genres.Zettel {
		if err = store.zettelIdIndex.AddZettelId(&child.ObjectId); err != nil {
			if errors.Is(err, object_id_provider.ErrDoesNotExist{}) {
				ui.Log().Printf("object id does not contain value: %s", err)
				err = nil
			} else {
				err = errors.Wrapf(err, "failed to write zettel to index: %s", child)
				return
			}
		}
	}

	if options.AddToInventoryList {
		ui.Log().Print("adding to inventory list", options, child)
		if err = store.commitTransacted(child, parent); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sku.String(child))
			return
		}
	}

	if options.AddToInventoryList ||
		options.StreamIndexOptions.AddToStreamIndex {
		if err = store.GetStreamIndex().Add(
			child,
			options,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if parent == nil {
			if child.GetGenre() == genres.Zettel {
				// TODO if this is a local zettel (i.e., not a different repo
				// and not a
				// different domain)

				// TODO verify that the zettel id consists of our identifiers,
				// otherwise
				// abort
			}

			if err = store.ui.TransactedNew(child); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			// [are/kabuto !task project-2021-zit-features zz-inbox] add delta
			// printing to changed objects
			// if err = s.Updated(mutter); err != nil {
			// 	err = errors.Wrap(err)
			// 	return
			// }

			if err = store.ui.TransactedUpdated(child); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	}

	if options.MergeCheckedOut {
		if err = store.ReadExternalAndMergeIfNecessary(
			child,
			parent,
			options,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (s *Store) fetchParentIfNecessary(
	sk *sku.Transacted,
) (mutter *sku.Transacted, err error) {
	mutter = sku.GetTransactedPool().Get()
	// TODO find a way to make this more performant when operating over sshfs
	if err = s.GetStreamIndex().ReadOneObjectId(
		sk.GetObjectId(),
		mutter,
	); err != nil {
		if collections.IsErrNotFound(err) || errors.IsNotExist(err) {
			// TODO decide if this should continue to virtual stores
			sku.GetTransactedPool().Put(mutter)
			mutter = nil
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	sk.Metadata.Mutter().ResetWith(mutter.Metadata.Sha())

	return
}

// TODO add results for which stores had which change types
func (store *Store) commitTransacted(
	object *sku.Transacted,
	parent *sku.Transacted,
) (err error) {
	if !store.inventoryList.LastTai.Less(object.GetTai()) {
		object.Metadata.Tai = store.GetTai()
	}

	if err = store.inventoryListStore.AddObjectToOpenList(
		store.inventoryList,
		object,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) handleUnchanged(
	t *sku.Transacted,
) (err error) {
	return s.ui.TransactedUnchanged(t)
}

func (s *Store) UpdateKonfig(
	sh interfaces.Sha,
) (kt *sku.Transacted, err error) {
	return s.CreateOrUpdateBlobSha(
		&ids.Config{},
		sh,
	)
}

func (s *Store) createTagsOrType(k *ids.ObjectId) (err error) {
	t := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(t)

	switch k.GetGenre() {
	default:
		err = genres.MakeErrUnsupportedGenre(k.GetGenre())
		return

	case genres.Type:
		t.GetMetadata().Type = ids.DefaultOrPanic(genres.Type)

	case genres.Tag:
	}

	if err = t.ObjectId.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.addObjectToAbbrStore(t); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.Commit(
		t,
		sku.CommitOptions{
			StoreOptions:       sku.GetStoreOptionsUpdate(),
			DontAddMissingTags: true,
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) addType(
	t ids.Type,
) (err error) {
	if t.IsEmpty() {
		err = errors.ErrorWithStackf("attempting to add empty type")
		return
	}

	var oid ids.ObjectId

	if err = oid.ResetWithIdLike(t); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.GetStreamIndex().ObjectExists(&oid); err == nil {
		return
	}

	err = nil

	var k ids.ObjectId

	if err = k.SetWithIdLike(t); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.createTagsOrType(&k); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) addTypeAndExpandedIfNecessary(
	rootTipe ids.Type,
) (err error) {
	if rootTipe.IsEmpty() {
		return
	}

	if ids.IsBuiltin(rootTipe) {
		return
	}

	typesExpanded := ids.ExpandOneSlice(
		rootTipe,
		ids.MakeType,
		expansion.ExpanderRight,
	)

	for _, tipe := range typesExpanded {
		if err = s.addType(tipe); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (s *Store) addTags(
	tags []ids.Tag,
) (err error) {
	s.tagLock.Lock()
	defer s.tagLock.Unlock()

	var oid ids.ObjectId

	for _, tag := range tags {
		if tag.IsVirtual() {
			continue
		}

		if err = oid.ResetWithIdLike(tag); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = s.GetStreamIndex().ObjectExists(&oid); err == nil {
			continue
		}

		err = nil

		var k ids.ObjectId

		if err = k.SetWithIdLike(tag); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = s.createTagsOrType(&k); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (s *Store) addMissingTypeAndTags(
	commitOptions sku.CommitOptions,
	object *sku.Transacted,
) (err error) {
	tipe := object.GetType()

	if !commitOptions.DontAddMissingType {
		if err = s.addTypeAndExpandedIfNecessary(tipe); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !commitOptions.DontAddMissingTags && object.GetGenre() == genres.Tag {
		var tag ids.Tag

		if err = tag.TodoSetFromObjectId(object.GetObjectId()); err != nil {
			err = errors.Wrap(err)
			return
		}

		tagsExpanded := ids.ExpandOneSlice(
			tag,
			ids.MakeTag,
			expansion.ExpanderRight,
		)

		if len(tagsExpanded) > 0 {
			tagsExpanded = tagsExpanded[:len(tagsExpanded)-1]
		}

		if err = s.addTags(tagsExpanded); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !commitOptions.DontAddMissingTags {
		es := quiter.SortedValues(object.Metadata.GetTags())

		if object.GetGenre() == genres.Tag {
			var tag ids.Tag

			if err = tag.TodoSetFromObjectId(object.GetObjectId()); err != nil {
				err = errors.Wrap(err)
				return
			}

			tagsExpanded := ids.ExpandOneSlice(
				tag,
				ids.MakeTag,
				expansion.ExpanderRight,
			)

			if len(tagsExpanded) > 0 {
				tagsExpanded = tagsExpanded[:len(tagsExpanded)-1]
			}

			if err = s.addTags(tagsExpanded); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		for _, e := range es {
			tagsExpanded := ids.ExpandOneSlice(
				e,
				ids.MakeTag,
				expansion.ExpanderRight,
			)

			if err = s.addTags(tagsExpanded); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	return
}

func (s *Store) addObjectToAbbrStore(m *sku.Transacted) (err error) {
	if err = s.GetAbbrStore().AddObjectToAbbreviationStore(m); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) reindexOne(object sku.ObjectWithList) (err error) {
	o := sku.CommitOptions{
		StoreOptions: sku.GetStoreOptionsReindex(),
	}

	if err = s.Commit(object.Object, o); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = s.GetAbbrStore().AddObjectToAbbreviationStore(
		object.Object,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
