package env_dir

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/id"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

type Mover struct {
	file *os.File
	*writer

	basePath                  string
	objectPath                string
	lockFile                  bool
	errorOnAttemptedOverwrite bool
}

func NewMover(moveOptions MoveOptions) (mover *Mover, err error) {
	mover = &Mover{
		lockFile:                  moveOptions.LockInternalFiles,
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

	writeOptions := WriteOptions{
		Config: moveOptions.Config,
		Writer: mover.file,
	}

	if mover.writer, err = NewWriter(writeOptions); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (m *Mover) Close() (err error) {
	if m.file == nil {
		err = errors.ErrorWithStackf("nil file")
		return
	}

	if m.writer == nil {
		err = errors.ErrorWithStackf("nil object reader")
		return
	}

	if err = m.writer.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// var fi os.FileInfo

	// if fi, err = m.file.Stat(); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	if err = files.Close(m.file); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh := m.GetShaLike()

	// log.Log().Printf(
	// 	"wrote %d bytes to %s, sha %s",
	// 	fi.Size(),
	// 	m.file.Name(),
	// 	sh,
	// )

	if m.objectPath == "" {
		// TODO-P3 move this validation to options
		if m.basePath == "" {
			err = errors.ErrorWithStackf("basepath is nil")
			return
		}

		if m.objectPath, err = id.MakeDirIfNecessary(sh, m.basePath); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	p := m.file.Name()

	if err = os.Rename(p, m.objectPath); err != nil {
		if files.Exists(m.objectPath) {
			if m.errorOnAttemptedOverwrite {
				err = MakeErrAlreadyExists(sh, m.objectPath)
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

	if m.lockFile {
		if err = files.SetDisallowUserChanges(m.objectPath); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
