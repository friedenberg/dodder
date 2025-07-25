package store_fs

import (
	"fmt"
	"os"
	"path"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) CheckoutOne(
	options checkout_options.Options,
	sz sku.TransactedGetter,
) (col sku.SkuType, err error) {
	col, _, err = store.checkoutOneIfNecessary(options, sz)
	return
}

func (store *Store) checkoutOneForReal(
	options checkout_options.Options,
	co *sku.CheckedOut,
	item *sku.FSItem,
) (err error) {
	if store.config.IsDryRun() {
		return
	}

	fsOptions := GetCheckoutOptionsFromOptions(options)

	// delete the existing checkout if it exists in the cwd
	if fsOptions.Path == PathOptionDefault {
		if err = store.RemoveItem(item); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var info checkoutFileNameInfo

	if err = store.hydrateCheckoutFileNameInfoFromCheckedOut(
		options,
		co,
		&info,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.setObjectIfNecessary(options, item, info); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.setBlobIfNecessary(
		options,
		item,
		info,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	// This is necessary otherwise External is an empty sku
	sku.Resetter.ResetWith(co.GetSkuExternal(), co.GetSku())

	if err = store.WriteFSItemToExternal(item, co.GetSkuExternal()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.fileEncoder.Encode(
		fsOptions.TextFormatterOptions,
		co.GetSkuExternal(),
		item,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) setObjectIfNecessary(
	options checkout_options.Options,
	i *sku.FSItem,
	info checkoutFileNameInfo,
) (err error) {
	if !options.CheckoutMode.IncludesMetadata() {
		i.MutableSetLike.Del(&i.Object)
		i.Object.Reset()
		return
	}

	if err = i.Object.SetPath(info.objectName); err != nil {
		err = errors.Wrap(err)
		return
	}

	i.MutableSetLike.Add(&i.Object)

	return
}

func (store *Store) setBlobIfNecessary(
	options checkout_options.Options,
	i *sku.FSItem,
	info checkoutFileNameInfo,
) (err error) {
	fsOptions := GetCheckoutOptionsFromOptions(options)

	if fsOptions.ForceInlineBlob ||
		!options.CheckoutMode.IncludesBlob() {
		i.MutableSetLike.Del(&i.Blob)
		i.Blob.Reset()
		return
	}

	fe := store.config.GetTypeExtension(info.tipe.String())

	if fe == "" {
		fe = info.tipe.StringSansOp()
	}

	if err = i.Blob.SetPath(
		info.basename + "." + fe,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	i.MutableSetLike.Add(&i.Blob)

	return
}

func (store *Store) shouldCheckOut(
	options checkout_options.Options,
	cz *sku.CheckedOut,
	allowMutterMatch bool,
) bool {
	if options.Force {
		return true
	}

	eq := object_metadata.EqualerSansTai.Equals(
		&cz.GetSku().Metadata,
		&cz.GetSkuExternal().Metadata,
	)

	if eq {
		return true
	}

	if !allowMutterMatch {
		ui.Log().Print("")
		return false
	}

	mutter := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(mutter)

	if err := store.storeSupplies.ReadOneInto(
		cz.GetSku().GetObjectId(),
		mutter,
	); err == nil {
		if object_metadata.EqualerSansTai.Equals(
			&mutter.Metadata,
			&cz.GetSkuExternal().Metadata,
		) {
			return true
		}
	}

	ui.Log().Print("")

	return false
}

type checkoutFileNameInfo struct {
	basename   string
	objectName string
	tipe       ids.Type
	inlineBlob bool
}

func (store *Store) hydrateCheckoutFileNameInfoFromCheckedOut(
	options checkout_options.Options,
	co *sku.CheckedOut,
	info *checkoutFileNameInfo,
) (err error) {
	if err = store.SetFilenameForTransacted(options, co.GetSku(), info); err != nil {
		err = errors.Wrap(err)
		return
	}

	co.SetState(checked_out_state.JustCheckedOut)

	info.tipe = co.GetSku().GetType()
	info.inlineBlob = store.config.IsInlineType(info.tipe)

	return
}

func (store *Store) SetFilenameForTransacted(
	options checkout_options.Options,
	sk *sku.Transacted,
	info *checkoutFileNameInfo,
) (err error) {
	cwd := store.envRepo.GetCwd()

	fsOptions := GetCheckoutOptionsFromOptions(options)

	if fsOptions.Path == PathOptionTempLocal {
		var f *os.File

		if f, err = store.envRepo.GetTempLocal().FileTempWithTemplate(
			fmt.Sprintf(
				"*.%s",
				store.FileExtensionForObject(sk),
			),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, f)

		info.basename = f.Name()
		info.objectName = f.Name()

		return
	}

	if sk.GetGenre() == genres.Zettel {
		var zettelId ids.ZettelId

		if err = zettelId.Set(sk.GetObjectId().String()); err != nil {
			err = errors.Wrap(err)
			return
		}

		if info.basename, err = env_dir.MakeDirIfNecessaryForStringerWithHeadAndTail(
			zettelId,
			cwd,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		info.objectName = store.PathForTransacted(cwd, sk)
	} else {
		info.basename = store.PathForTransacted(cwd, sk)
		info.objectName = info.basename
	}

	if strings.Contains(info.basename, "!") {
		err = errors.ErrorWithStackf(
			"contains illegal characters: %q",
			info.basename,
		)
		return
	}

	if strings.Contains(info.objectName, "!") {
		err = errors.ErrorWithStackf(
			"contains illegal characters: %q",
			info.objectName,
		)
		return
	}

	return
}

func (store *Store) PathForTransacted(dir string, sk *sku.Transacted) string {
	return path.Join(
		dir,
		fmt.Sprintf(
			"%s.%s",
			sk.GetObjectId().StringSansOp(),
			store.FileExtensionForObject(sk),
		),
	)
}

func (store *Store) FileExtensionForObject(
	sk *sku.Transacted,
) string {
	var extension string

	if sk.GetGenre() == genres.Blob {
		extension = store.config.GetTypeExtension(sk.GetType().String())

		if extension == "" {
			extension = sk.GetType().StringSansOp()
		}
	} else {
		extension = store.fileExtensions.GetFileExtensionForGenre(sk)
	}

	if extension == "" {
		extension = "unknown"
	}

	return extension
}

func (store *Store) RemoveItem(i *sku.FSItem) (err error) {
	// TODO check conflict state
	for fdItem := range i.MutableSetLike.All() {
		if err = fdItem.Remove(store.envRepo); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	i.Reset()

	return
}

func (store *Store) UpdateCheckoutFromCheckedOut(
	options checkout_options.OptionsWithoutMode,
	co sku.SkuType,
) (err error) {
	o := checkout_options.Options{
		OptionsWithoutMode: options,
	}

	if o.CheckoutMode, err = store.GetCheckoutMode(
		co.GetSkuExternal(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if o.CheckoutMode == checkout_mode.None {
		return
	}

	fsOptions := GetCheckoutOptionsFromOptionsWithoutMode(options)
	fsOptions.Path = PathOptionTempLocal
	options.StoreSpecificOptions = fsOptions

	var replacement *sku.CheckedOut
	var oldFDs, newFDs *sku.FSItem

	if oldFDs, err = store.ReadFSItemFromExternal(co.GetSkuExternal()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if replacement, newFDs, err = store.checkoutOneIfNecessary(
		o,
		co.GetSkuExternal(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer GetCheckedOutPool().Put(replacement)

	if !oldFDs.Object.IsEmpty() &&
		!newFDs.Object.IsEmpty() &&
		!store.config.IsDryRun() {
		if err = os.Rename(
			newFDs.Object.GetPath(),
			oldFDs.Object.GetPath(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !oldFDs.Blob.IsEmpty() &&
		!newFDs.Blob.IsEmpty() &&
		!store.config.IsDryRun() {
		if err = os.Rename(
			newFDs.Blob.GetPath(),
			oldFDs.Blob.GetPath(),
		); err != nil {
			err = errors.Wrapf(
				err,
				"New: %q, Old: %q",
				newFDs.Blob.GetPath(),
				oldFDs.Blob.GetPath(),
			)

			return
		}
	}

	return
}
