package store_config

import (
	"encoding/gob"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_blobs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

func init() {
	gob.Register(repo_config_blobs.V1{})
	gob.Register(repo_config_blobs.V0{})
}

func (store *store) recompile(
	blobStore typed_blob_store.Stores,
) (err error) {
	if err = store.recompileTags(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.recompileTypes(blobStore); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *store) recompileTags() (err error) {
	store.ImplicitTags = make(implicitTagMap)

	if err = store.compiled.Tags.Each(
		func(ke *tag) (err error) {
			var e ids.Tag

			if err = e.Set(ke.String()); err != nil {
				err = errors.Wrapf(err, "Sku: %s", sku.StringTaiGenreObjectIdShaBlob(&ke.Transacted))
				return
			}

			if err = store.AccumulateImplicitTags(e); err != nil {
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

func (store *store) recompileTypes(
	blobStore typed_blob_store.Stores,
) (err error) {
	inlineTypes := collections_value.MakeMutableValueSet[values.String](nil)

	defer func() {
		store.InlineTypes = inlineTypes.CloneSetLike()
	}()

	if err = store.Types.Each(
		func(ct *sku.Transacted) (err error) {
			tipe := ct.GetSku().GetType()
			var commonBlob type_blobs.Blob

			if commonBlob, _, err = blobStore.Type.ParseTypedBlob(
				tipe,
				ct.GetBlobSha(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			defer blobStore.Type.PutTypedBlob(tipe, commonBlob)

			if commonBlob == nil {
				err = errors.ErrorWithStackf("nil type blob for type: %q. Sku: %s", tipe, ct)
				return
			}

			fe := commonBlob.GetFileExtension()

			if fe == "" {
				fe = ct.GetObjectId().StringSansOp()
			}

			// TODO-P2 enforce uniqueness
			store.ExtensionsToTypes[fe] = ct.GetObjectId().String()
			store.TypesToExtensions[ct.GetObjectId().String()] = fe

			isBinary := commonBlob.GetBinary()
			if !isBinary {
				inlineTypes.Add(values.MakeString(ct.ObjectId.String()))
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}
	return
}

func (store *store) HasChanges() (ok bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	ok = len(store.compiled.changes) > 0

	if ok {
		ui.Log().Print(store.compiled.changes)
	}

	return
}

func (store *store) GetChanges() (out []string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	out = make([]string, len(store.changes))
	copy(out, store.changes)

	return
}

func (kc *compiled) SetNeedsRecompile(reason string) {
	kc.lock.Lock()
	defer kc.lock.Unlock()

	kc.setNeedsRecompile(reason)
}

func (kc *compiled) setNeedsRecompile(reason string) {
	kc.changes = append(kc.changes, reason)
}

func (store *store) loadMutableConfig(
	envRepo env_repo.Env,
) (err error) {
	var file *os.File

	p := envRepo.FileConfigMutable()

	if file, err = files.Open(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	dec := gob.NewDecoder(file)

	if err = dec.Decode(&store.compiled); err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	// TODO replace with triple_hyphen_io
	if err = store.loadMutableConfigBlob(
		store.Sku.GetType(),
		store.Sku.GetBlobSha(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *store) Flush(
	dirLayout env_repo.Env,
	blobStore typed_blob_store.Stores,
	printerHeader interfaces.FuncIter[string],
) (err error) {
	if !store.HasChanges() || store.IsDryRun() {
		return
	}

	wg := errors.MakeWaitGroupParallel()
	wg.Do(func() (err error) {
		if err = store.flushMutableConfig(dirLayout, blobStore, printerHeader); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	})

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.changes = store.changes[:0]

	return
}

func (store *store) flushMutableConfig(
	s env_repo.Env,
	blobStore typed_blob_store.Stores,
	printerHeader interfaces.FuncIter[string],
) (err error) {
	if err = printerHeader("recompiling konfig"); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.recompile(blobStore); err != nil {
		err = errors.Wrap(err)
		return
	}

	p := s.FileConfigMutable()

	var f *os.File

	if f, err = files.OpenCreateWriteOnlyTruncate(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	enc := gob.NewEncoder(f)

	if err = enc.Encode(&store.compiled); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = printerHeader("recompiled konfig"); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
