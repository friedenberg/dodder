package store_fs

import (
	"encoding/gob"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/delta/file_extensions"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

func init() {
	gob.Register(sku.Transacted{})
}

func Make(
	config sku.Config,
	deletedPrinter interfaces.FuncIter[*fd.FD],
	fileExtensions file_extensions.Config,
	envRepo env_repo.Env,
) (fs *Store, err error) {
	fs = &Store{
		config:         config,
		deletedPrinter: deletedPrinter,
		envRepo:        envRepo,
		fileEncoder:    MakeFileEncoder(envRepo, config),
		fileExtensions: fileExtensions,
		dir:            envRepo.GetCwd(),
		dirInfo: makeObjectsWithDir(
			fileExtensions,
			envRepo,
		),
		deleted: collections_value.MakeMutableValueSet[*fd.FD](
			nil,
		),
		deletedInternal: collections_value.MakeMutableValueSet[*fd.FD](
			nil,
		),
		metadataTextParser: object_metadata.MakeTextParser(
			object_metadata.Dependencies{
				EnvDir:         envRepo,
				BlobStore:      envRepo.GetDefaultBlobStore(),
				BlobDigestType: envRepo.GetConfigPublic().Blob.GetBlobDigestTypeString(),
			},
		),
	}

	return
}

type Store struct {
	config             sku.Config
	deletedPrinter     interfaces.FuncIter[*fd.FD]
	metadataTextParser object_metadata.TextParser
	envRepo            env_repo.Env
	fileEncoder        FileEncoder
	inlineTypeChecker  ids.InlineTypeChecker
	fileExtensions     file_extensions.Config
	dir                string

	dirInfo

	deleteLock      sync.Mutex
	deleted         fd.MutableSet
	deletedInternal fd.MutableSet
}

func (store *Store) GetExternalStoreLike() store_workspace.StoreLike {
	return store
}

// Deletions of user objects that should be exposed to the user
func (store *Store) DeleteCheckedOut(co *sku.CheckedOut) (err error) {
	external := co.GetSkuExternal()

	var item *sku.FSItem

	if item, err = store.ReadFSItemFromExternal(external); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.deleteLock.Lock()
	defer store.deleteLock.Unlock()

	for fd := range item.FDs.All() {
		store.deleted.Add(fd)
	}

	return
}

// Deletions of "transient" internal objects that should not be exposed to the
// user
func (store *Store) DeleteCheckedOutInternal(co *sku.CheckedOut) (err error) {
	external := co.GetSkuExternal()

	var i *sku.FSItem

	if i, err = store.ReadFSItemFromExternal(external); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.deleteLock.Lock()
	defer store.deleteLock.Unlock()

	for fd := range i.FDs.All() {
		store.deletedInternal.Add(fd)
	}

	return
}

func (store *Store) Flush() (err error) {
	deleteOp := DeleteCheckout{}

	if err = deleteOp.Run(
		store.config.IsDryRun(),
		store.envRepo,
		store.deletedPrinter,
		store.deleted,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = deleteOp.Run(
		store.config.IsDryRun(),
		store.envRepo,
		nil,
		store.deletedInternal,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.deleted.Reset()
	store.deletedInternal.Reset()

	return
}

func (store *Store) String() (out string) {
	if quiter.Len(
		store.dirInfo.probablyCheckedOut,
		store.definitelyNotCheckedOut,
	) == 0 {
		return
	}

	sb := &strings.Builder{}
	sb.WriteRune(doddish.OpGroupOpen)

	hasOne := false

	writeOneIfNecessary := func(v interfaces.Stringer) (err error) {
		if hasOne {
			sb.WriteRune(doddish.OpOr)
		}

		sb.WriteString(v.String())

		hasOne = true

		return
	}

	for z := range store.dirInfo.probablyCheckedOut.All() {
		writeOneIfNecessary(z)
	}

	for z := range store.definitelyNotCheckedOut.All() {
		writeOneIfNecessary(z)
	}

	sb.WriteRune(doddish.OpGroupClose)

	out = sb.String()
	return
}

func (store *Store) GetExternalObjectIds() (ks []*sku.FSItem, err error) {
	if err = store.dirInfo.processRootDir(); err != nil {
		err = errors.Wrap(err)
		return
	}

	ks = make([]*sku.FSItem, 0)
	var l sync.Mutex

	if err = store.All(
		func(kfp *sku.FSItem) (err error) {
			l.Lock()
			defer l.Unlock()

			ks = append(ks, kfp)

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) GetFSItemsForDir(
	fd *fd.FD,
) (items []*sku.FSItem, err error) {
	if !fd.IsDir() {
		err = errors.ErrorWithStackf("not a directory: %q", fd)
		return
	}

	if items, err = store.dirInfo.processDir(fd.GetPath()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO confirm against actual Object Id
func (store *Store) GetFSItemsForString(
	baseDir string,
	value string,
	tryPattern bool,
) (items []*sku.FSItem, err error) {
	if value == "." {
		if items, err = store.GetExternalObjectIds(); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}

	var fdee *fd.FD

	if fdee, err = fd.MakeFromPath(
		baseDir,
		value,
		store.envRepo.GetDefaultBlobStore(),
	); err != nil {
		if errors.IsNotExist(err) && tryPattern {
			if items, err = store.dirInfo.processFDPattern(
				value,
				filepath.Join(store.dir, fmt.Sprintf("%s*", value)),
				store.dir,
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	if fdee.IsDir() {
		if items, err = store.GetFSItemsForDir(fdee); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if _, items, err = store.dirInfo.processFD(fdee); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (store *Store) GetObjectIdsForString(
	queryLiteral string,
) (objectIds []sku.ExternalObjectId, err error) {
	var items []*sku.FSItem

	if items, err = store.GetFSItemsForString(
		store.root,
		queryLiteral,
		false,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, item := range items {
		var eoid ids.ExternalObjectId

		if err = item.WriteToExternalObjectId(
			&eoid,
			store.envRepo,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		objectIds = append(objectIds, &eoid)
	}

	return
}

func (store *Store) Get(
	k interfaces.ObjectId,
) (t *sku.FSItem, ok bool) {
	return store.dirInfo.probablyCheckedOut.Get(k.String())
}

func (store *Store) Initialize(
	storeSupplies store_workspace.Supplies,
) (err error) {
	store.root = storeSupplies.WorkspaceDir
	store.storeSupplies = storeSupplies
	return
}

func (store *Store) ReadAllExternalItems() (err error) {
	if err = store.dirInfo.processRootDir(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) ReadFSItemFromExternal(
	tg sku.TransactedGetter,
) (item *sku.FSItem, err error) {
	item = &sku.FSItem{} // TODO use pool or use dir_items?
	item.Reset()

	sk := tg.GetSku()

	// TODO handle sort order
	for _, field := range sk.Metadata.Fields {
		var fdee *fd.FD

		switch strings.ToLower(field.Key) {
		case "object":
			fdee = &item.Object

		case "blob":
			fdee = &item.Blob

		case "conflict":
			fdee = &item.Conflict

		default:
			err = errors.ErrorWithStackf("unexpected field: %#v", field)
			return
		}

		// if we've already set one of object, blob, or conflict, don't set it
		// again
		// and instead add a new FD to the item
		if !fdee.IsEmpty() {
			fdee = &fd.FD{}
		}

		if err = fdee.SetIgnoreNotExists(field.Value); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = item.FDs.Add(fdee); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = item.ExternalObjectId.SetObjectIdLike(
		&sk.ObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	// external.ObjectId.ResetWith(conflicted.GetSkuExternal().GetObjectId())
	// TODO populate FD
	if !sk.ExternalObjectId.IsEmpty() {
		if err = item.ExternalObjectId.SetObjectIdLike(
			&item.ExternalObjectId,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (store *Store) WriteFSItemToExternal(
	item *sku.FSItem,
	tg sku.TransactedGetter,
) (err error) {
	external := tg.GetSku()
	external.Metadata.Fields = external.Metadata.Fields[:0]

	m := &external.Metadata
	m.Tai = item.GetTai()

	var mode checkout_mode.Mode

	if mode, err = item.GetCheckoutModeOrError(); err != nil {
		if sku.IsErrMergeConflict(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	switch mode {
	case checkout_mode.BlobOnly:
		before := item.Blob.String()
		after := store.envRepo.Rel(before)

		if err = external.ExternalObjectId.SetBlob(after); err != nil {
			err = errors.Wrap(err)
			return
		}

	default:
		k := &item.ExternalObjectId

		if err = external.ObjectId.SetObjectIdLike(k); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = external.ExternalObjectId.SetObjectIdLike(
			&item.ExternalObjectId,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if external.ExternalObjectId.String() != k.String() {
			err = errors.ErrorWithStackf(
				"expected %q but got %q. %s",
				k,
				&external.ExternalObjectId,
				item.Debug(),
			)

			return
		}
	}

	if err = item.WriteToSku(
		external,
		store.envRepo,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	fdees := quiter.SortedValues(item.FDs)

	for _, f := range fdees {
		field := object_metadata.Field{
			Value:     f.GetPath(),
			ColorType: string_format_writer.ColorTypeId,
		}

		switch {
		case item.Object.Equals(f):
			field.Key = "object"

		case item.Conflict.Equals(f):
			field.Key = "conflict"

		case item.Blob.Equals(f):
			fallthrough

		default:
			field.Key = "blob"
		}

		external.Metadata.Fields = append(external.Metadata.Fields, field)
	}

	return
}
