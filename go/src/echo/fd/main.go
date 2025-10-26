package fd

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/thyme"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
)

type FD struct {
	isDir   bool
	path    string
	modTime thyme.Time
	digest  markl.Id
	state   State
}

func (a *FD) IsStdin() bool {
	return a.path == "-"
}

func (a *FD) ModTime() thyme.Time {
	return a.modTime
}

func (a *FD) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a *FD) Equals2(b *FD) (bool, string) {
	if a.path != b.path {
		return false, "path"
	}

	// if !a.modTime.Equals(b.modTime) {
	// 	return false, "modTime"
	// }

	// if !a.sha.Equals(&b.sha) {
	// 	return false, "sha"
	// }

	return true, ""
}

func (a *FD) Equals(b *FD) bool {
	if a.path != b.path {
		return false
	}

	if !a.modTime.Equals(b.modTime) {
		return false
	}

	if !markl.Equals(&a.digest, &b.digest) {
		return false
	}

	return true
}

func (fd *FD) SetFromPath(
	baseDir string,
	path string,
	blobStore interfaces.BlobWriterFactory,
) (err error) {
	if path == "" {
		err = errors.ErrorWithStackf("nil file desriptor")
		return err
	}

	if path == "." {
		err = errors.ErrorWithStackf("'.' not supported")
		return err
	}

	if !filepath.IsAbs(path) {
		path = filepath.Clean(filepath.Join(baseDir, path))
	}

	var fileInfo os.FileInfo

	if fileInfo, err = os.Stat(path); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = fd.SetFromFileInfoWithDir(
		fileInfo,
		filepath.Dir(path),
		blobStore,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (fd *FD) SetFromFileInfoWithDir(
	fileInfo os.FileInfo,
	dir string,
	blobStore interfaces.BlobWriterFactory,
) (err error) {
	if err = fd.SetFileInfoWithDir(fileInfo, dir); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if fileInfo.IsDir() {
		return err
	}

	// TODO eventually enforce requirement of blob writer factory
	if blobStore == nil {
		return err
	}

	var file *os.File

	if file, err = files.OpenExclusiveReadOnly(fd.GetPath()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, file)

	var writer interfaces.BlobWriter

	if writer, err = blobStore.MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, writer)

	if _, err = io.Copy(writer, file); err != nil {
		err = errors.Wrap(err)
		return err
	}

	markl.SetDigester(&fd.digest, writer)
	fd.state = StateStored

	return err
}

func (fd *FD) SetWithBlobWriterFactory(
	path string,
	blobStore interfaces.BlobWriterFactory,
) (err error) {
	if path == "" {
		err = errors.ErrorWithStackf("empty path")
		return err
	}

	if blobStore == nil {
		panic("BlobWriterFactory is nil")
	}

	var f *os.File

	if f, err = files.OpenExclusiveReadOnly(path); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, f)

	var blobWriter interfaces.BlobWriter

	if blobWriter, err = blobStore.MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, blobWriter)

	if _, err = io.Copy(blobWriter, f); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var fi os.FileInfo

	if fi, err = f.Stat(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = fd.SetFileInfoWithDir(fi, filepath.Dir(path)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	fd.path = path
	markl.SetDigester(&fd.digest, blobWriter)
	fd.state = StateStored

	return err
}

func (f *FD) SetFileInfoWithDir(fi os.FileInfo, dir string) (err error) {
	f.Reset()
	f.isDir = fi.IsDir()
	f.modTime = thyme.Tyme(fi.ModTime())

	p := dir

	if !f.isDir {
		p = filepath.Join(dir, fi.Name())
	}

	if f.path, err = filepath.Abs(p); err != nil {
		err = errors.Wrap(err)
		return err
	}

	f.state = StateFileInfo

	return err
}

func (fd *FD) SetIgnoreNotExists(v string) (err error) {
	v = strings.TrimSpace(v)

	if v == "-" {
		fd.path = v
		fd.modTime = thyme.Now()
		fd.isDir = false
		return err
	}

	if v == "." {
		err = errors.ErrorWithStackf("'.' not supported")
		return err
	}

	fd.path = filepath.Clean(v)
	fd.state = StateFileInfo

	return err
}

func (fd *FD) Set(v string) (err error) {
	v = strings.TrimSpace(v)

	if v == "-" {
		fd.path = v
		fd.modTime = thyme.Now()
		fd.isDir = false
		return err
	}

	if v == "." {
		err = errors.ErrorWithStackf("'.' not supported")
		return err
	}

	var fi os.FileInfo

	if fi, err = os.Stat(v); err != nil {
		err = errors.Wrapf(err, "Value: %q", v)
		return err
	}

	if err = fd.SetFileInfoWithDir(fi, filepath.Dir(v)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (f *FD) String() string {
	p := filepath.Clean(f.path)

	if f.isDir {
		return p + string(filepath.Separator)
	} else {
		return p
	}
}

func (f *FD) DepthRelativeTo(dir string) int {
	dir = filepath.Clean(dir)
	rel, err := filepath.Rel(dir, f.GetPath())

	if err != nil || strings.HasPrefix(rel, "..") {
		return -1
	}

	return strings.Count(rel, string(filepath.Separator))
}

func (e *FD) Ext() string {
	return Ext(e.path)
}

func (e *FD) ExtSansDot() string {
	return ExtSansDot(e.path)
}

func (e *FD) FilePathSansExt() string {
	base := e.path
	ext := e.Ext()
	return base[:len(base)-len(ext)]
}

func (e *FD) FileName() string {
	return filepath.Base(e.path)
}

func (e *FD) FileNameSansExt() string {
	return FileNameSansExt(e.path)
}

func (e *FD) FileNameSansExtRelTo(d string) (string, error) {
	return FileNameSansExtRelTo(e.path, d)
}

func (e *FD) FilePathRelTo(d string) (string, error) {
	rel, err := filepath.Rel(d, e.path)
	if err != nil {
		return e.path, nil
		// return "", err
	}

	return rel, nil
}

func (e *FD) DirBaseOnly() string {
	return DirBaseOnly(e.path)
}

func (f *FD) IsEmpty() bool {
	if f.path == "" {
		return true
	}

	// if f.ModTime.IsZero() {
	// 	return true
	// }

	return false
}

func (fd *FD) Parts() [3]string {
	return [3]string{"", "", fd.String()}
}

func (fd *FD) GetPath() string {
	return fd.path
}

func (fd *FD) SetPath(p string) (err error) {
	fd.path = p
	return err
}

func (fd *FD) SetPathRel(p, dir string) (err error) {
	if p, err = filepath.Rel(dir, p); err != nil {
		err = errors.Wrapf(err, "Name: %q, Dir: %q", p, dir)
		return err
	}

	if err = fd.SetPath(p); err != nil {
		err = errors.Wrapf(err, "Name: %q, Dir: %q", p, dir)
		return err
	}

	return err
}

func (fd *FD) IsDir() bool {
	return fd.isDir
}

func (fd *FD) SetShaLike(v interfaces.MarklId) (err error) {
	return fd.digest.SetDigest(v)
}

func (fd *FD) GetDigest() interfaces.MarklId {
	return &fd.digest
}

func (fd *FD) GetState() State {
	return fd.state
}

func (fd *FD) Exists() bool {
	if fd.path == "" {
		return false
	}

	return files.Exists(fd.path)
}

func (fd *FD) Remove(directoryLayout directory_layout.Mutable) (err error) {
	if fd.path == "" {
		return err
	}

	if err = directoryLayout.Delete(fd.path); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	return err
}

func (fd *FD) Reset() {
	fd.state = StateUnknown
	fd.isDir = false
	fd.path = ""
	fd.modTime.Reset()
	fd.digest.Reset()
}

func (dst *FD) ResetWith(src *FD) {
	dst.state = src.state
	dst.isDir = src.isDir
	dst.path = src.path
	dst.modTime = src.modTime
	dst.digest.ResetWith(src.digest)
}

func (src *FD) Clone() (dst *FD) {
	dst = &FD{}
	dst.state = src.state
	dst.isDir = src.isDir
	dst.path = src.path
	dst.modTime = src.modTime
	dst.digest.ResetWith(src.digest)
	return dst
}
