package blob_stores

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type remoteSftp struct {
	uiPrinter ui.Printer

	buckets []int

	config blob_store_configs.ConfigSFTPRemotePath

	multiHash       bool
	defaultHashType markl.HashType

	// TODO populate blobIOWrapper with env_repo.FileNameBlobStoreConfig at
	// `config.GetRemotePath()`
	blobIOWrapper interfaces.BlobIOWrapper
	sshClient     *ssh.Client
	sftpClient    *sftp.Client

	// TODO extract below into separate struct
	blobCacheLock sync.RWMutex
	blobCache     map[string]struct{}
}

var _ interfaces.BlobStore = &remoteSftp{}

func makeSftpStore(
	ctx interfaces.ActiveContext,
	uiPrinter ui.Printer,
	config blob_store_configs.ConfigSFTPRemotePath,
	sshClient *ssh.Client,
) (blobStore *remoteSftp, err error) {
	var defaultHashType markl.HashType

	if defaultHashType, err = markl.GetHashTypeOrError(
		defaultHashTypeId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	blobStore = &remoteSftp{
		defaultHashType: defaultHashType,
		uiPrinter:       uiPrinter,
		buckets:         defaultBuckets,
		config:          config,
		sshClient:       sshClient,
		blobCache:       make(map[string]struct{}),
	}

	ui.Log().Print("creating sftp client")

	if blobStore.sftpClient, err = sftp.NewClient(blobStore.sshClient); err != nil {
		err = errors.Wrapf(err, "failed to create SFTP client")
		return
	}

	ctx.After(errors.MakeFuncContextFromFuncErr(blobStore.close))

	if err = blobStore.initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore *remoteSftp) GetBlobStoreConfig() interfaces.BlobStoreConfig {
	return blobStore.config
}

func (blobStore *remoteSftp) GetDefaultHashType() interfaces.HashType {
	return blobStore.defaultHashType
}

func (blobStore *remoteSftp) close() (err error) {
	if blobStore.sftpClient != nil {
		if err = blobStore.sftpClient.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return nil
}

func (blobStore *remoteSftp) initialize() (err error) {
	remotePath := blobStore.config.GetRemotePath()
	// TODO read remote blob store config (including hash buckets)

	// Create directory tree if it doesn't exist
	parts := strings.Split(remotePath, "/")
	var currentPath string

	for _, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" && !strings.HasPrefix(remotePath, "/") {
			currentPath = part
		} else {
			currentPath = path.Join(currentPath, part)
		}

		blobStore.uiPrinter.Printf("checking directory %q...", currentPath)
		if _, err = blobStore.sftpClient.Stat(currentPath); err != nil {
			// TODO check error
			blobStore.uiPrinter.Printf("creating directory %q...", currentPath)
			if err = blobStore.sftpClient.Mkdir(currentPath); err != nil {
				// Directory might exist, continue
				err = nil
			}
		}
	}

	return
}

func (blobStore *remoteSftp) GetBlobStoreDescription() string {
	return fmt.Sprintf("TODO: sftp")
}

func (blobStore *remoteSftp) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	return blobStore.blobIOWrapper
}

func (blobStore *remoteSftp) GetLocalBlobStore() interfaces.BlobStore {
	return blobStore
}

func (blobStore *remoteSftp) makeEnvDirConfig() env_dir.Config {
	return env_dir.DefaultConfig
	// return env_dir.MakeConfig(
	// 	blobStore.blobIOWrapper.GetBlobCompression(),
	// 	blobStore.blobIOWrapper.GetBlobEncryption(),
	// )
}

func (blobStore *remoteSftp) remotePathForMerkleId(
	merkleId interfaces.MarklId,
) string {
	return env_dir.MakeHashBucketPathFromMerkleId(
		merkleId,
		blobStore.buckets,
		strings.TrimPrefix(blobStore.config.GetRemotePath(), "/"),
	)
}

func (blobStore *remoteSftp) HasBlob(
	merkleId interfaces.MarklId,
) (ok bool) {
	if merkleId.IsNull() {
		ok = true
		return
	}

	blobStore.blobCacheLock.RLock()

	if _, ok = blobStore.blobCache[string(merkleId.GetBytes())]; ok {
		blobStore.blobCacheLock.RUnlock()
		return
	}

	blobStore.blobCacheLock.RUnlock()

	remotePath := blobStore.remotePathForMerkleId(merkleId)

	if _, err := blobStore.sftpClient.Stat(remotePath); err == nil {
		blobStore.blobCacheLock.Lock()
		blobStore.blobCache[string(merkleId.GetBytes())] = struct{}{}
		blobStore.blobCacheLock.Unlock()
		ok = true
	}

	return
}

func (blobStore *remoteSftp) AllBlobs() interfaces.SeqError[interfaces.MarklId] {
	return func(yield func(interfaces.MarklId, error) bool) {
		basePath := strings.TrimPrefix(blobStore.config.GetRemotePath(), "/")

		// Walk through the two-level directory structure (Git-like bucketing)
		walker := blobStore.sftpClient.Walk(basePath)

		digest, repool := blobStore.defaultHashType.GetBlobId()
		defer repool()

		for walker.Step() {
			if err := walker.Err(); err != nil {
				if !yield(nil, errors.Wrapf(err, "BasePath: %q", basePath)) {
					return
				}

				continue
			}

			info := walker.Stat()

			if info.IsDir() {
				continue
			}

			relPath := strings.TrimPrefix(walker.Path(), basePath)
			relPath = strings.TrimPrefix(relPath, "/")

			if err := markl.SetHexStringFromPath(digest, relPath); err != nil {
				if !yield(nil, errors.Wrap(err)) {
					return
				}

				continue
			}

			blobStore.blobCacheLock.Lock()
			blobStore.blobCache[string(digest.GetBytes())] = struct{}{}
			blobStore.blobCacheLock.Unlock()

			if !yield(digest, nil) {
				return
			}
		}
	}
}

func (blobStore *remoteSftp) BlobWriter(
	marklHashTypeId string,
) (w interfaces.WriteCloseMarklIdGetter, err error) {
	mover := &sftpMover{
		store:  blobStore,
		config: blobStore.makeEnvDirConfig(),
	}

	if err = mover.initialize(blobStore.defaultHashType.Get()); err != nil {
		err = errors.Wrap(err)
		return
	}

	w = mover
	return
}

func (blobStore *remoteSftp) Mover() (mover interfaces.Mover, err error) {
	return blobStore.BlobWriter("")
}

func (blobStore *remoteSftp) BlobReader(
	digest interfaces.MarklId,
) (readCloser interfaces.ReadCloseMarklIdGetter, err error) {
	if digest.IsNull() {
		readCloser = markl_io.MakeNopReadCloser(
			blobStore.defaultHashType.Get(),
			io.NopCloser(bytes.NewReader(nil)),
		)
		return
	}

	remotePath := blobStore.remotePathForMerkleId(digest)

	var remoteFile *sftp.File

	if remoteFile, err = blobStore.sftpClient.Open(remotePath); err != nil {
		if os.IsNotExist(err) {
			err = env_dir.ErrBlobMissing{
				BlobId: markl.Clone(digest),
				Path:   remotePath,
			}
		} else {
			err = errors.Wrap(err)
		}
		return
	}

	blobStore.blobCacheLock.Lock()
	blobStore.blobCache[string(digest.GetBytes())] = struct{}{}
	blobStore.blobCacheLock.Unlock()

	// Create streaming reader that handles decompression/decryption
	config := blobStore.makeEnvDirConfig()
	streamingReader := &sftpStreamingReader{
		file:   remoteFile,
		config: config,
	}

	if readCloser, err = streamingReader.createReader(
		blobStore.defaultHashType.Get(),
	); err != nil {
		remoteFile.Close()
		err = errors.Wrap(err)
		return
	}

	return
}

// sftpMover implements interfaces.Mover and interfaces.ShaWriteCloser
// TODO explore using env_dir.Mover generically instead of this
type sftpMover struct {
	hash     interfaces.Hash
	store    *remoteSftp
	config   env_dir.Config
	tempFile *sftp.File
	tempPath string
	writer   *sftpWriter
	closed   bool
}

func (mover *sftpMover) initialize(hash interfaces.Hash) (err error) {
	mover.hash = hash

	// Create a temporary file on the remote server
	var tempNameBytes [16]byte
	if _, err = rand.Read(tempNameBytes[:]); err != nil {
		err = errors.Wrap(err)
		return
	}

	tempName := fmt.Sprintf("tmp_%x", tempNameBytes)
	mover.tempPath = path.Join(mover.store.config.GetRemotePath(), tempName)

	if mover.tempFile, err = mover.store.sftpClient.Create(mover.tempPath); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Create the streaming writer with compression/encryption

	if mover.writer, err = newSftpWriter(
		mover.config,
		mover.tempFile,
		hash,
	); err != nil {
		mover.tempFile.Close()
		mover.store.sftpClient.Remove(mover.tempPath)
		err = errors.Wrap(err)
		return
	}

	return
}

func (mover *sftpMover) Write(p []byte) (n int, err error) {
	if mover.writer == nil {
		err = errors.ErrorWithStackf("writer not initialized")
		return
	}

	return mover.writer.Write(p)
}

func (mover *sftpMover) ReadFrom(r io.Reader) (n int64, err error) {
	if mover.writer == nil {
		err = errors.ErrorWithStackf("writer not initialized")
		return
	}

	return mover.writer.ReadFrom(r)
}

func (mover *sftpMover) Close() (err error) {
	if mover.closed {
		return nil
	}

	mover.closed = true

	// Ensure cleanup happens
	// TODO capture errors using errors.Deferred*
	defer func() {
		if mover.tempFile != nil {
			mover.tempFile.Close()
		}
		if mover.tempPath != "" {
			mover.store.sftpClient.Remove(mover.tempPath)
		}
	}()

	if mover.writer == nil {
		// No data was written
		return nil
	}

	// Close the writer to finalize compression/encryption and digest
	// calculation
	if err = mover.writer.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Close the temp file
	if err = mover.tempFile.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Get the calculated digest and determine final path
	blobDigest := mover.writer.GetDigest()
	finalPath := mover.store.remotePathForMerkleId(blobDigest)

	// Ensure the target directory exists (Git-like bucketing)
	finalDir := path.Dir(finalPath)
	if err = mover.store.sftpClient.MkdirAll(finalDir); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Atomically move temp file to final location
	if err = mover.store.sftpClient.Rename(mover.tempPath, finalPath); err != nil {
		// Check if file already exists
		if _, statErr := mover.store.sftpClient.Stat(finalPath); statErr == nil {
			// File already exists, this is OK - just remove temp file
			mover.store.sftpClient.Remove(mover.tempPath)
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	mover.store.blobCacheLock.Lock()
	mover.store.blobCache[string(blobDigest.GetBytes())] = struct{}{}
	mover.store.blobCacheLock.Unlock()

	// Clear temp path so cleanup doesn't try to remove it
	mover.tempPath = ""

	return
}

func (mover *sftpMover) GetMarklId() interfaces.MarklId {
	if mover.writer == nil {
		return mover.GetMarklId()
	}

	return mover.writer.GetDigest()
}

// sftpWriter implements the streaming writer with compression/encryption
type sftpWriter struct {
	hash            interfaces.Hash
	tee             io.Writer
	wCompress, wAge io.WriteCloser
	wBuf            *bufio.Writer
}

func newSftpWriter(
	config env_dir.Config,
	ioWriter io.Writer,
	hash interfaces.Hash,
) (writer *sftpWriter, err error) {
	writer = &sftpWriter{}

	writer.wBuf = bufio.NewWriter(ioWriter)

	if writer.wAge, err = config.GetBlobEncryption().WrapWriter(writer.wBuf); err != nil {
		err = errors.Wrap(err)
		return
	}

	writer.hash = hash

	if writer.wCompress, err = config.GetBlobCompression().WrapWriter(writer.wAge); err != nil {
		err = errors.Wrap(err)
		return
	}

	writer.tee = io.MultiWriter(writer.hash, writer.wCompress)

	return
}

func (writer *sftpWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if n, err = io.Copy(writer.tee, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (writer *sftpWriter) Write(p []byte) (n int, err error) {
	return writer.tee.Write(p)
}

func (writer *sftpWriter) Close() (err error) {
	if writer.wCompress != nil {
		if err = writer.wCompress.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if writer.wAge != nil {
		if err = writer.wAge.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if writer.wBuf != nil {
		if err = writer.wBuf.Flush(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (writer *sftpWriter) GetDigest() interfaces.MarklId {
	id, _ := writer.hash.GetMarklId()
	return id
}

// sftpStreamingReader handles decompression/decryption while reading from SFTP
// TODO combine with sftpReader
type sftpStreamingReader struct {
	file   *sftp.File
	config interfaces.BlobIOWrapper
}

func (reader *sftpStreamingReader) createReader(
	hash interfaces.Hash,
) (readCloser interfaces.ReadCloseMarklIdGetter, err error) {
	// Create streaming reader with decompression/decryption
	sftpReader := &sftpReader{
		file:   reader.file,
		config: reader.config,
	}

	if err = sftpReader.initialize(hash); err != nil {
		err = errors.Wrap(err)
		return
	}

	readCloser = sftpReader

	return
}

// sftpReader implements streaming decompression/decryption for SFTP
type sftpReader struct {
	file      *sftp.File
	config    interfaces.BlobIOWrapper
	hash      interfaces.Hash
	decrypter io.Reader
	expander  io.ReadCloser
	tee       io.Reader
}

func (reader *sftpReader) initialize(hash interfaces.Hash) (err error) {
	// Set up decryption
	if reader.decrypter, err = reader.config.GetBlobEncryption().WrapReader(reader.file); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Set up decompression
	if reader.expander, err = reader.config.GetBlobCompression().WrapReader(reader.decrypter); err != nil {
		err = errors.Wrap(err)
		return
	}

	reader.hash = hash
	reader.tee = io.TeeReader(reader.expander, reader.hash)

	return
}

func (reader *sftpReader) Read(p []byte) (n int, err error) {
	return reader.tee.Read(p)
}

func (reader *sftpReader) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, reader.tee)
}

func (reader *sftpReader) Seek(
	offset int64,
	whence int,
) (actual int64, err error) {
	seeker, ok := reader.decrypter.(io.Seeker)

	if !ok {
		err = errors.ErrorWithStackf("seeking not supported")
		return
	}

	return seeker.Seek(offset, whence)
}

func (reader *sftpReader) Close() error {
	// TODO capture both errors using errors.Join
	err1 := reader.expander.Close()
	err2 := reader.file.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

func (reader *sftpReader) GetMarklId() interfaces.MarklId {
	id, _ := reader.hash.GetMarklId()
	return id
}
