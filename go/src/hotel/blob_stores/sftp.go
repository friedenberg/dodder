package blob_stores

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/id"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type sftpConfig interface {
	GetHost() string
	GetPort() int
	GetUser() string
	GetPassword() string
	GetPrivateKeyPath() string
	GetRemotePath() string
	GetLockInternalFiles() bool
}

type sftpBlobStore struct {
	config sftpConfig

	// TODO populate blobIOWrapper with env_repo.FileNameBlobStoreConfig at
	// `config.GetRemotePath()`
	blobIOWrapper interfaces.BlobIOWrapper
	sshClient     *ssh.Client
	sftpClient    *sftp.Client
}

func makeSftpStore(
	ctx errors.Context,
	config sftpConfig,
) (store *sftpBlobStore, err error) {
	store = &sftpBlobStore{
		config: config,
	}

	if err = store.connect(); err != nil {
		err = errors.Wrap(err)
		return
	}

	ctx.After(store.close)

	if err = store.ensureRemotePath(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO review this
func (blobStore *sftpBlobStore) connect() (err error) {
	sshConfig := &ssh.ClientConfig{
		User:            blobStore.config.GetUser(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: make this configurable
	}

	// Configure authentication
	if blobStore.config.GetPrivateKeyPath() != "" {
		var key ssh.Signer
		var keyBytes []byte

		if keyBytes, err = os.ReadFile(blobStore.config.GetPrivateKeyPath()); err != nil {
			err = errors.Wrapf(err, "failed to read private key")
			return
		}

		if key, err = ssh.ParsePrivateKey(keyBytes); err != nil {
			err = errors.Wrapf(err, "failed to parse private key")
			return
		}

		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(key)}
	} else if blobStore.config.GetPassword() != "" {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(blobStore.config.GetPassword())}
	} else {
		err = errors.Errorf("no authentication method configured")
		return
	}

	addr := fmt.Sprintf(
		"%s:%d",
		blobStore.config.GetHost(),
		blobStore.config.GetPort(),
	)

	if blobStore.sshClient, err = ssh.Dial("tcp", addr, sshConfig); err != nil {
		err = errors.Wrapf(err, "failed to connect to SSH server")
		return
	}

	if blobStore.sftpClient, err = sftp.NewClient(blobStore.sshClient); err != nil {
		err = errors.Wrapf(err, "failed to create SFTP client")
		blobStore.sshClient.Close()
		return
	}

	return
}

// TODO determine how these errors should or should not cascade
func (blobStore *sftpBlobStore) close() (err error) {
	if blobStore.sftpClient != nil {
		if err = blobStore.sftpClient.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if blobStore.sshClient != nil {
		if err = blobStore.sshClient.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return nil
}

func (blobStore *sftpBlobStore) ensureRemotePath() (err error) {
	remotePath := blobStore.config.GetRemotePath()
	// TODO read remote blob store config

	// Create directory tree if it doesn't exist
	parts := strings.Split(remotePath, "/")
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		if currentPath == "" && !strings.HasPrefix(remotePath, "/") {
			currentPath = part
		} else {
			currentPath = path.Join(currentPath, part)
		}

		if _, err = blobStore.sftpClient.Stat(currentPath); err != nil {
			if err = blobStore.sftpClient.Mkdir(currentPath); err != nil {
				// Directory might exist, continue
				err = nil
			}
		}
	}

	return
}

func (blobStore *sftpBlobStore) GetBlobStoreDescription() string {
	return fmt.Sprintf("TODO: sftp")
}

func (blobStore *sftpBlobStore) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	return blobStore.blobIOWrapper
}

func (blobStore *sftpBlobStore) GetLocalBlobStore() interfaces.LocalBlobStore {
	return blobStore
}

func (blobStore *sftpBlobStore) makeEnvDirConfig() env_dir.Config {
	return env_dir.MakeConfig(
		blobStore.blobIOWrapper.GetBlobCompression(),
		blobStore.blobIOWrapper.GetBlobEncryption(),
	)
}

func (blobStore *sftpBlobStore) remotePathForSha(sh interfaces.Sha) string {
	return path.Join(
		blobStore.config.GetRemotePath(),
		id.Path(sh.GetShaLike(), ""),
	)
}

func (blobStore *sftpBlobStore) HasBlob(sh interfaces.Sha) (ok bool) {
	if sh.GetShaLike().IsNull() {
		ok = true
		return
	}

	remotePath := blobStore.remotePathForSha(sh)
	if _, err := blobStore.sftpClient.Stat(remotePath); err == nil {
		ok = true
	}

	return
}

func (blobStore *sftpBlobStore) AllBlobs() interfaces.SeqError[interfaces.Sha] {
	return func(yield func(interfaces.Sha, error) bool) {
		basePath := blobStore.config.GetRemotePath()

		// Walk through the two-level directory structure (Git-like bucketing)
		walker := blobStore.sftpClient.Walk(basePath)

		for walker.Step() {
			if err := walker.Err(); err != nil {
				if !yield(nil, errors.Wrap(err)) {
					return
				}
			}

			info := walker.Stat()
			if info.IsDir() {
				continue
			}

			// TODO replace with id.Path
			// Extract SHA from path
			relPath := strings.TrimPrefix(walker.Path(), basePath)
			relPath = strings.TrimPrefix(relPath, "/")

			// Skip if not in expected format (2 chars / remaining chars)
			parts := strings.Split(relPath, "/")
			if len(parts) != 2 || len(parts[0]) != 2 {
				continue
			}

			shaStr := parts[0] + parts[1]
			var sh sha.Sha

			if err := sh.Set(shaStr); err != nil {
				if !yield(nil, errors.Wrap(err)) {
					return
				}
				continue
			}

			if !yield(&sh, nil) {
				return
			}
		}
	}
}

func (blobStore *sftpBlobStore) BlobWriter() (w interfaces.ShaWriteCloser, err error) {
	mover := &sftpMover{
		store:  blobStore,
		config: blobStore.makeEnvDirConfig(),
	}

	if err = mover.initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	w = mover
	return
}

func (blobStore *sftpBlobStore) Mover() (mover interfaces.Mover, err error) {
	return blobStore.BlobWriter()
}

func (blobStore *sftpBlobStore) BlobReader(
	sh interfaces.Sha,
) (r interfaces.ShaReadCloser, err error) {
	if sh.GetShaLike().IsNull() {
		r = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	remotePath := blobStore.remotePathForSha(sh)

	var remoteFile *sftp.File

	if remoteFile, err = blobStore.sftpClient.Open(remotePath); err != nil {
		if os.IsNotExist(err) {
			shCopy := sha.GetPool().Get()
			shCopy.ResetWithShaLike(sh.GetShaLike())

			err = env_dir.ErrBlobMissing{
				ShaGetter: shCopy,
				Path:      remotePath,
			}
		} else {
			err = errors.Wrap(err)
		}
		return
	}

	// Create streaming reader that handles decompression/decryption
	config := blobStore.makeEnvDirConfig()
	streamingReader := &sftpStreamingReader{
		file:   remoteFile,
		config: config,
	}

	if r, err = streamingReader.createReader(); err != nil {
		remoteFile.Close()
		err = errors.Wrap(err)
		return
	}

	return
}

// sftpMover implements interfaces.Mover and interfaces.ShaWriteCloser
// TODO explore using env_dir.Mover generically instead of this
type sftpMover struct {
	store    *sftpBlobStore
	config   env_dir.Config
	tempFile *sftp.File
	tempPath string
	writer   *sftpWriter
	closed   bool
}

func (mover *sftpMover) initialize() (err error) {
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

	// Close the writer to finalize compression/encryption and SHA calculation
	if err = mover.writer.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Close the temp file
	if err = mover.tempFile.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Get the calculated SHA and determine final path
	sh := mover.writer.GetShaLike()
	finalPath := mover.store.remotePathForSha(sh)

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

	// Clear temp path so cleanup doesn't try to remove it
	mover.tempPath = ""

	return
}

func (mover *sftpMover) GetShaLike() interfaces.Sha {
	if mover.writer == nil {
		// Return empty SHA if no data written
		// TODO use sha.GetPool()
		return &sha.Sha{}
	}

	return mover.writer.GetShaLike()
}

// sftpWriter implements the streaming writer with compression/encryption
type sftpWriter struct {
	hash            hash.Hash
	tee             io.Writer
	wCompress, wAge io.WriteCloser
	wBuf            *bufio.Writer
}

func newSftpWriter(
	config env_dir.Config,
	ioWriter io.Writer,
) (writer *sftpWriter, err error) {
	writer = &sftpWriter{}

	writer.wBuf = bufio.NewWriter(ioWriter)

	if writer.wAge, err = config.GetBlobEncryption().WrapWriter(writer.wBuf); err != nil {
		err = errors.Wrap(err)
		return
	}

	writer.hash = sha256.New()

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

func (writer *sftpWriter) GetShaLike() interfaces.Sha {
	return sha.FromHash(writer.hash)
}

// sftpStreamingReader handles decompression/decryption while reading from SFTP
// TODO combine with sftpReader
type sftpStreamingReader struct {
	file   *sftp.File
	config interfaces.BlobIOWrapper
}

func (reader *sftpStreamingReader) createReader() (readCloser interfaces.ShaReadCloser, err error) {
	// Create streaming reader with decompression/decryption
	sftpReader := &sftpReader{
		file:   reader.file,
		config: reader.config,
	}

	if err = sftpReader.initialize(); err != nil {
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
	hash      hash.Hash
	decrypter io.Reader
	expander  io.ReadCloser
	tee       io.Reader
}

func (reader *sftpReader) initialize() (err error) {
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

	// Set up SHA calculation
	// TODO sha.GetPool()
	reader.hash = sha256.New()
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

func (reader *sftpReader) GetShaLike() interfaces.Sha {
	return sha.FromHash(reader.hash)
}
