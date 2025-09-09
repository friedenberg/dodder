package env_dir

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

// TODO fold into markl_io
type MoveOptions struct {
	TemporaryFS
	ErrorOnAttemptedOverwrite bool
	FinalPath                 string
	GenerateFinalPathFromSha  bool
}

type localFileMover struct {
	funcJoin func(string, ...string) string
	file     *os.File
	interfaces.BlobWriter

	basePath                  string
	objectPath                string
	lockFile                  bool
	errorOnAttemptedOverwrite bool
}

func NewMover(
	config Config,
	moveOptions MoveOptions,
) (interfaces.BlobWriter, error) {
	// TODO make MoveOptions an interface and add support for localFileShaMover
	// and localFinalPathMover
	return newMover(config, moveOptions)
}

// TODO add back support for locking internal files
// TODO split mover into sha-based mover and final-path based mover
// TODO extract writer portion in injected depenency
func newMover(
	config Config,
	moveOptions MoveOptions,
) (mover *localFileMover, err error) {
	mover = &localFileMover{
		funcJoin:                  config.funcJoin,
		errorOnAttemptedOverwrite: moveOptions.ErrorOnAttemptedOverwrite,
	}

	if moveOptions.GenerateFinalPathFromSha {
		mover.basePath = moveOptions.FinalPath
	} else {
		mover.objectPath = moveOptions.FinalPath
	}

	if mover.file, err = moveOptions.FileTemp(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if mover.BlobWriter, err = NewWriter(
		config,
		mover.file,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (mover *localFileMover) Close() (err error) {
	if mover.file == nil {
		err = errors.ErrorWithStackf("nil file")
		return
	}

	if mover.BlobWriter == nil {
		err = errors.ErrorWithStackf("nil object reader")
		return
	}

	if err = mover.BlobWriter.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// var fi os.FileInfo

	// if fi, err = m.file.Stat(); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	if err = mover.file.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	digest := mover.GetMarklId()

	if err = markl.MakeErrEmptyType(digest); err != nil {
		err = errors.Wrap(err)
		return
	}

	if digest.IsNull() {
		return
	}

	// if err = merkle.MakeErrIsNull(digest, ""); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	// log.Log().Printf(
	// 	"wrote %d bytes to %s, sha %s",
	// 	fi.Size(),
	// 	m.file.Name(),
	// 	sh,
	// )

	if mover.objectPath == "" {
		// TODO-P3 move this validation to options
		if mover.basePath == "" {
			err = errors.ErrorWithStackf("basepath is nil")
			return
		}

		if mover.objectPath, err = MakeDirIfNecessary(
			markl.Format(digest),
			mover.funcJoin,
			mover.basePath,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	path := mover.file.Name()

	if err = os.Rename(path, mover.objectPath); err != nil {
		if files.Exists(mover.objectPath) {
			if mover.errorOnAttemptedOverwrite {
				err = MakeErrBlobAlreadyExists(digest, mover.objectPath)
			} else {
				err = nil
			}

			return
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	// log.Log().Printf("moved %s to %s", p, m.objectPath)

	if mover.lockFile {
		if err = files.SetDisallowUserChanges(mover.objectPath); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
