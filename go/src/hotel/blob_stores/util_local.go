package blob_stores

import (
	"io/fs"
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

		if err := filepath.WalkDir(
			basePath,
			func(path string, dirEntry fs.DirEntry, in error) (err error) {
				if in != nil {
					err = errors.Wrap(in)
					return
				}

				if path == basePath {
					return
				}

				if dirEntry.IsDir() {
					return
				}

				if err = markl.SetHexStringFromAbsolutePath(id, path, basePath); err != nil {
					if !yield(nil, errors.Wrap(err)) {
						err = filepath.SkipAll
						return
					}

					return
				}

				if id.IsNull() {
					return
				}

				if !yield(id, nil) {
					err = filepath.SkipAll
					return
				}

				return
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
