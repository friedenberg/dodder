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

func (kc *store) recompile(
	blobStore typed_blob_store.Stores,
) (err error) {
	if err = kc.recompileTags(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = kc.recompileTypes(blobStore); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (kc *store) recompileTags() (err error) {
	kc.ImplicitTags = make(implicitTagMap)

	if err = kc.compiled.Tags.Each(
		func(ke *tag) (err error) {
			var e ids.Tag

			if err = e.Set(ke.String()); err != nil {
				err = errors.Wrapf(err, "Sku: %s", sku.StringTaiGenreObjectIdShaBlob(&ke.Transacted))
				return
			}

			if err = kc.AccumulateImplicitTags(e); err != nil {
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

func (kc *store) recompileTypes(
	blobStore typed_blob_store.Stores,
) (err error) {
	inlineTypes := collections_value.MakeMutableValueSet[values.String](nil)

	defer func() {
		kc.InlineTypes = inlineTypes.CloneSetLike()
	}()

	if err = kc.Types.Each(
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
			kc.ExtensionsToTypes[fe] = ct.GetObjectId().String()
			kc.TypesToExtensions[ct.GetObjectId().String()] = fe

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

func (kc *store) HasChanges() (ok bool) {
	kc.lock.Lock()
	defer kc.lock.Unlock()

	ok = len(kc.compiled.changes) > 0

	if ok {
		ui.Log().Print(kc.compiled.changes)
	}

	return
}

func (kc *store) GetChanges() (out []string) {
	kc.lock.Lock()
	defer kc.lock.Unlock()

	out = make([]string, len(kc.changes))
	copy(out, kc.changes)

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

func (kc *store) loadMutableConfig(
	dirLayout env_repo.Env,
) (err error) {
	var f *os.File

	p := dirLayout.FileConfigMutable()

	if f, err = files.Open(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	dec := gob.NewDecoder(f)

	if err = dec.Decode(&kc.compiled); err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	if err = kc.loadMutableConfigBlob(
		kc.Sku.GetType(),
		kc.Sku.GetBlobSha(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (kc *store) Flush(
	dirLayout env_repo.Env,
	blobStore typed_blob_store.Stores,
	printerHeader interfaces.FuncIter[string],
) (err error) {
	if !kc.HasChanges() || kc.IsDryRun() {
		return
	}

	wg := errors.MakeWaitGroupParallel()
	wg.Do(func() (err error) {
		if err = kc.flushMutableConfig(dirLayout, blobStore, printerHeader); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	})

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	kc.changes = kc.changes[:0]

	return
}

func (kc *store) flushMutableConfig(
	s env_repo.Env,
	blobStore typed_blob_store.Stores,
	printerHeader interfaces.FuncIter[string],
) (err error) {
	if err = printerHeader("recompiling konfig"); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = kc.recompile(blobStore); err != nil {
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

	if err = enc.Encode(&kc.compiled); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = printerHeader("recompiled konfig"); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
