package fd

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

func MakeFromDirPath(
	path string,
) (fd *FD, err error) {
	fd = &FD{}
	fd.isDir = true
	fd.path = path

	return
}

func MakeFromPathAndDirEntry(
	path string,
	dirEntry fs.DirEntry,
	blobWriter interfaces.BlobWriterFactory,
) (fd *FD, err error) {
	if path == "" {
		err = errors.ErrorWithStackf("nil file desriptor")
		return
	}

	if path == "." {
		err = errors.ErrorWithStackf("'.' not supported")
		return
	}

	var fi os.FileInfo

	if fi, err = dirEntry.Info(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if fd, err = MakeFromFileInfoWithDir(fi, filepath.Dir(path), blobWriter); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MakeFromPath(
	baseDir string,
	path string,
	blobWriter interfaces.BlobWriterFactory,
) (fd *FD, err error) {
	if path == "" {
		err = errors.ErrorWithStackf("nil file desriptor")
		return
	}

	if path == "." {
		err = errors.ErrorWithStackf("'.' not supported")
		return
	}

	if !filepath.IsAbs(path) {
		path = filepath.Clean(filepath.Join(baseDir, path))
	}

	var fi os.FileInfo

	if fi, err = os.Stat(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	if fd, err = MakeFromFileInfoWithDir(
		fi,
		filepath.Dir(path),
		blobWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MakeFromFileInfoWithDir(
	fileInfo os.FileInfo,
	dir string,
	blobStore interfaces.BlobWriterFactory,
) (fd *FD, err error) {
	// TODO use pool
	fd = &FD{}

	if err = fd.SetFileInfoWithDir(fileInfo, dir); err != nil {
		err = errors.Wrap(err)
		return
	}

	if fileInfo.IsDir() {
		return
	}

	// TODO eventually enforce requirement of blob writer factory
	if blobStore == nil {
		return
	}

	var file *os.File

	if file, err = files.OpenExclusiveReadOnly(fd.GetPath()); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	var writer interfaces.WriteCloseMarklIdGetter

	if writer, err = blobStore.MakeBlobWriter(""); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writer)

	if _, err = io.Copy(writer, file); err != nil {
		err = errors.Wrap(err)
		return
	}

	markl.SetDigester(&fd.digest, writer)
	fd.state = StateStored

	return
}
