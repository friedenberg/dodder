package store_config

import (
	"encoding/gob"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

func init() {
	gob.Register(repo_configs.V1{})
	gob.Register(repo_configs.V0{})
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
	store.config.ImplicitTags = make(implicitTagMap)

	for ke := range store.config.Tags.All() {
		var e ids.Tag

		if err = e.Set(ke.String()); err != nil {
			err = errors.Wrapf(
				err,
				"Sku: %s",
				sku.StringTaiGenreObjectIdShaBlob(&ke.Transacted),
			)
			return
		}

		if err = store.config.AccumulateImplicitTags(e); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (store *store) recompileTypes(
	blobStore typed_blob_store.Stores,
) (err error) {
	inlineTypes := collections_value.MakeMutableValueSet[values.String](nil)

	defer func() {
		store.config.InlineTypes = inlineTypes.CloneSetLike()
	}()

	for ct := range store.config.Types.All() {
		tipe := ct.GetSku().GetType()
		var commonBlob type_blobs.Blob
		var repool interfaces.FuncRepool

		if commonBlob, repool, _, err = blobStore.Type.ParseTypedBlob(
			tipe,
			ct.GetBlobDigest(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer repool()

		if commonBlob == nil {
			err = errors.ErrorWithStackf(
				"nil type blob for type: %q. Sku: %s",
				tipe,
				ct,
			)
			return
		}

		fe := commonBlob.GetFileExtension()

		if fe == "" {
			fe = ct.GetObjectId().StringSansOp()
		}

		// TODO-P2 enforce uniqueness
		store.config.ExtensionsToTypes[fe] = ct.GetObjectId().String()
		store.config.TypesToExtensions[ct.GetObjectId().String()] = fe

		isBinary := commonBlob.GetBinary()
		if !isBinary {
			inlineTypes.Add(values.MakeString(ct.ObjectId.String()))
		}

	}
	return
}

func (store *store) HasChanges() (ok bool) {
	store.config.lock.Lock()
	defer store.config.lock.Unlock()

	ok = len(store.config.compiled.changes) > 0

	if ok {
		ui.Log().Print(store.config.compiled.changes)
	}

	return
}

func (store *store) GetChanges() (out []string) {
	store.config.lock.Lock()
	defer store.config.lock.Unlock()

	out = make([]string, len(store.config.changes))
	copy(out, store.config.changes)

	return
}

func (compiled *compiled) SetNeedsRecompile(reason string) {
	compiled.lock.Lock()
	defer compiled.lock.Unlock()

	compiled.setNeedsRecompile(reason)
}

func (compiled *compiled) setNeedsRecompile(reason string) {
	compiled.changes = append(compiled.changes, reason)
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

	if err = dec.Decode(&store.config.compiled); err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	if err = store.loadMutableConfigBlob(
		store.config.Sku.GetType(),
		store.config.Sku.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *store) Flush(
	envRepo env_repo.Env,
	blobStore typed_blob_store.Stores,
	printerHeader interfaces.FuncIter[string],
) (err error) {
	if !store.HasChanges() || store.config.IsDryRun() {
		return
	}

	waitGroup := errors.MakeWaitGroupParallel()
	waitGroup.Do(func() (err error) {
		if err = store.flushMutableConfig(envRepo, blobStore, printerHeader); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	})

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.config.changes = store.config.changes[:0]

	return
}

func (store *store) flushMutableConfig(
	envRepo env_repo.Env,
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

	path := envRepo.FileConfigMutable()

	var file *os.File

	if file, err = files.OpenCreateWriteOnlyTruncate(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	enc := gob.NewEncoder(file)

	if err = enc.Encode(&store.config.compiled); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = printerHeader("recompiled konfig"); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *store) loadMutableConfigBlob(
	mutableConfigType ids.Type,
	blobSha interfaces.BlobId,
) (err error) {
	var readCloser io.ReadCloser

	if readCloser, err = store.envRepo.GetDefaultBlobStore().BlobReader(
		blobSha,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	typedBlob := repo_configs.TypedBlob{
		Type: mutableConfigType,
	}

	if _, err = repo_configs.Coder.DecodeFrom(
		&typedBlob,
		readCloser,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.config.configRepo = typedBlob.Blob

	return
}
