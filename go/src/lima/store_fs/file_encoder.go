package store_fs

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type FileEncoder interface {
	Encode(
		checkout_options.TextFormatterOptions,
		*sku.Transacted,
		*sku.FSItem,
	) error
}

type fileEncoder struct {
	mode              int
	perm              os.FileMode
	envRepo           env_repo.Env
	inlineTypeChecker ids.InlineTypeChecker

	object_metadata.TextFormatterFamily
}

func MakeFileEncoder(
	envRepo env_repo.Env,
	inlineTypeChecker ids.InlineTypeChecker,
) *fileEncoder {
	blobStore := envRepo.GetDefaultBlobStore()

	return &fileEncoder{
		mode:              os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
		perm:              0o666,
		envRepo:           envRepo,
		inlineTypeChecker: inlineTypeChecker,
		TextFormatterFamily: object_metadata.MakeTextFormatterFamily(
			object_metadata.Dependencies{
				EnvDir:    envRepo,
				BlobStore: blobStore,
			},
		),
	}
}

func (encoder *fileEncoder) openOrCreate(p string) (file *os.File, err error) {
	if file, err = files.OpenFile(p, encoder.mode, encoder.perm); err != nil {
		err = errors.Wrap(err)

		if errors.IsExist(err) {
			// err = nil
			var err2 error

			if file, err2 = files.OpenExclusiveReadOnly(p); err2 != nil {
				err = errors.Wrap(err2)
			}
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	return
}

func (encoder *fileEncoder) EncodeObject(
	options checkout_options.TextFormatterOptions,
	object *sku.Transacted,
	objectPath string,
	blobPath string,
) (err error) {
	ctx := object_metadata.TextFormatterContext{
		PersistentFormatterContext: object.GetSku(),
		TextFormatterOptions:       options,
	}

	inline := encoder.inlineTypeChecker.IsInlineType(object.GetType())

	var ar interfaces.ReadCloseMarklIdGetter

	if ar, err = encoder.envRepo.GetDefaultBlobStore().BlobReader(object.GetBlobDigest()); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, ar)

	switch {
	case blobPath != "" && objectPath != "":
		var fileBlob, fileObject *os.File

		{
			if fileBlob, err = encoder.openOrCreate(
				blobPath,
			); err != nil {
				if errors.IsExist(err) {
					var aw interfaces.WriteCloseMarklIdGetter

					if aw, err = encoder.envRepo.GetDefaultBlobStore().BlobWriter(""); err != nil {
						err = errors.Wrap(err)
						return
					}

					defer errors.DeferredCloser(&err, aw)

					if _, err = io.Copy(aw, fileBlob); err != nil {
						err = errors.Wrap(err)
						return
					}

				} else {
					err = errors.Wrap(err)
					return
				}
			}

			defer errors.DeferredCloser(&err, fileBlob)

			if _, err = io.Copy(fileBlob, ar); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		if fileObject, err = encoder.openOrCreate(
			objectPath,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, fileObject)

		if _, err = encoder.BlobPath.FormatMetadata(fileObject, ctx); err != nil {
			err = errors.Wrap(err)
			return
		}

	case blobPath != "":
		var fBlob *os.File

		if fBlob, err = encoder.openOrCreate(
			blobPath,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, fBlob)

		if _, err = io.Copy(fBlob, ar); err != nil {
			err = errors.Wrap(err)
			return
		}

	case objectPath != "":
		var mtw object_metadata.TextFormatter

		if inline {
			mtw = encoder.InlineBlob
		} else {
			mtw = encoder.MetadataOnly
		}

		var fZettel *os.File

		if fZettel, err = encoder.openOrCreate(
			objectPath,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, fZettel)

		if _, err = mtw.FormatMetadata(fZettel, ctx); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (encoder *fileEncoder) Encode(
	options checkout_options.TextFormatterOptions,
	object *sku.Transacted,
	fsItem *sku.FSItem,
) (err error) {
	return encoder.EncodeObject(
		options,
		object,
		fsItem.Object.GetPath(),
		fsItem.Blob.GetPath(),
	)
}
