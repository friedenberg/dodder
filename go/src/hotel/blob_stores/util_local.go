package blob_stores

import (
	"io/fs"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

func localAllBlobs(
	basePath string,
	defaultHashType markl.FormatHash,
) interfaces.SeqError[interfaces.MarklId] {
	return func(yield func(interfaces.MarklId, error) bool) {
		id, repool := defaultHashType.GetBlobId()
		defer repool()

		if basePath == "" {
			yield(nil, errors.Errorf("empty base path"))
			return
		}

		{
			var err error
			var newBasePath string

			if newBasePath, err = os.Readlink(basePath); err != nil {
				if errors.IsReadlinkInvalidArgument(err) {
					err = nil
				} else {
					yield(nil, errors.Wrap(err))
					return
				}
			} else {
				basePath = newBasePath
			}
		}

		if err := filepath.WalkDir(
			basePath,
			func(path string, dirEntry fs.DirEntry, in error) (err error) {
				if in != nil {
					err = errors.Wrapf(in, "BasePath: %q", basePath)
					return err
				}

				if path == basePath {
					return err
				}

				if dirEntry.IsDir() {
					return err
				}

				if err = markl.SetHexStringFromAbsolutePath(id, path, basePath); err != nil {
					if !yield(nil, errors.Wrap(err)) {
						if dirEntry.IsDir() {
							err = filepath.SkipDir
						} else {
							err = nil
						}

						return err
					}

					return err
				}

				if id.IsNull() {
					return err
				}

				if !yield(id, nil) {
					err = filepath.SkipAll
					return err
				}

				return err
			},
		); err != nil {
			if !yield(nil, errors.Wrap(err)) {
				return
			}
		}
	}
}

func localAllBlobsMultihash(
	basePath string,
) interfaces.SeqError[interfaces.MarklId] {
	return func(yield func(interfaces.MarklId, error) bool) {
		dirnames, err := files.DirNames(basePath)
		if err != nil {
			yield(nil, errors.Wrap(err))
			return
		}

		for _, dirname := range dirnames {
			hashTypeId := filepath.Base(dirname)

			if hashTypeId == "." {
				continue
			}

			hashType, err := markl.GetFormatHashOrError(hashTypeId)
			if err != nil {
				if !yield(nil, errors.Wrap(err)) {
					return
				}

				continue
			}

			seq := localAllBlobs(dirname, hashType)

			for id, err := range seq {
				if !yield(id, err) {
					return
				}
			}
		}
	}
}
