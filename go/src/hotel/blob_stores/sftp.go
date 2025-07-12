package blob_stores

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"iter"
	"os"
	"path"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/id"
	"code.linenisgreat.com/dodder/go/src/delta/compression_type"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type sftpConfig interface {
	blob_store_configs.Config
	GetHost() string
	GetPort() int
	GetUser() string
	GetPassword() string
	GetPrivateKeyPath() string
	GetRemotePath() string
	GetLockInternalFiles() bool
}

type sftpBlobStore struct {
	config     sftpConfig
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func makeSftpStore(
	config sftpConfig,
	tempFS env_dir.TemporaryFS,
) (store *sftpBlobStore, err error) {
	store = &sftpBlobStore{
		config: config,
	}

	if err = store.connect(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Ensure remote base path exists
	if err = store.ensureRemotePath(); err != nil {
		err = errors.Wrap(err)
		store.close()
		return nil, err
	}

	return
}

func (store *sftpBlobStore) connect() (err error) {
	sshConfig := &ssh.ClientConfig{
		User:            store.config.GetUser(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: make this configurable
	}

	// Configure authentication
	if store.config.GetPrivateKeyPath() != "" {
		var key ssh.Signer
		var keyBytes []byte

		if keyBytes, err = os.ReadFile(store.config.GetPrivateKeyPath()); err != nil {
			err = errors.Wrapf(err, "failed to read private key")
			return
		}

		if key, err = ssh.ParsePrivateKey(keyBytes); err != nil {
			err = errors.Wrapf(err, "failed to parse private key")
			return
		}

		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(key)}
	} else if store.config.GetPassword() != "" {
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(store.config.GetPassword())}
	} else {
		err = errors.Errorf("no authentication method configured")
		return
	}

	addr := fmt.Sprintf("%s:%d", store.config.GetHost(), store.config.GetPort())
	
	if store.sshClient, err = ssh.Dial("tcp", addr, sshConfig); err != nil {
		err = errors.Wrapf(err, "failed to connect to SSH server")
		return
	}

	if store.sftpClient, err = sftp.NewClient(store.sshClient); err != nil {
		err = errors.Wrapf(err, "failed to create SFTP client")
		store.sshClient.Close()
		return
	}

	return
}

func (store *sftpBlobStore) close() {
	if store.sftpClient != nil {
		store.sftpClient.Close()
	}
	if store.sshClient != nil {
		store.sshClient.Close()
	}
}

func (store *sftpBlobStore) ensureRemotePath() (err error) {
	remotePath := store.config.GetRemotePath()
	
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
		
		if _, err = store.sftpClient.Stat(currentPath); err != nil {
			if err = store.sftpClient.Mkdir(currentPath); err != nil {
				// Directory might exist, continue
				err = nil
			}
		}
	}
	
	return
}

func (store *sftpBlobStore) GetBlobStore() interfaces.BlobStore {
	return store
}

func (store *sftpBlobStore) GetLocalBlobStore() interfaces.LocalBlobStore {
	return store
}

func (store *sftpBlobStore) makeEnvDirConfig() env_dir.Config {
	return env_dir.Config{
		Compression:       store.config.GetBlobCompression(),
		Encryption:        store.config.GetBlobEncryption(),
		LockInternalFiles: store.config.GetLockInternalFiles(),
	}
}

func (store *sftpBlobStore) remotePathForSha(sh interfaces.Sha) string {
	return path.Join(store.config.GetRemotePath(), id.Path(sh.GetShaLike(), ""))
}

func (store *sftpBlobStore) HasBlob(sh interfaces.Sha) (ok bool) {
	if sh.GetShaLike().IsNull() {
		ok = true
		return
	}

	remotePath := store.remotePathForSha(sh)
	if _, err := store.sftpClient.Stat(remotePath); err == nil {
		ok = true
	}

	return
}

func (store *sftpBlobStore) AllBlobs() iter.Seq2[interfaces.Sha, error] {
	return func(yield func(interfaces.Sha, error) bool) {
		basePath := store.config.GetRemotePath()
		
		// Walk through the two-level directory structure (Git-like bucketing)
		walker := store.sftpClient.Walk(basePath)
		
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

func (store *sftpBlobStore) BlobWriter() (w interfaces.ShaWriteCloser, err error) {
	mover := &sftpMover{
		store:  store,
		config: store.makeEnvDirConfig(),
	}
	
	if err = mover.initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	w = mover
	return
}

func (store *sftpBlobStore) Mover() (mover interfaces.Mover, err error) {
	return store.BlobWriter()
}

func (store *sftpBlobStore) BlobReader(sh interfaces.Sha) (r interfaces.ShaReadCloser, err error) {
	if sh.GetShaLike().IsNull() {
		r = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	remotePath := store.remotePathForSha(sh)
	
	// Open remote file for reading
	var remoteFile *sftp.File
	if remoteFile, err = store.sftpClient.Open(remotePath); err != nil {
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
	config := store.makeEnvDirConfig()
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
type sftpMover struct {
	store      *sftpBlobStore
	config     env_dir.Config
	tempFile   *sftp.File
	tempPath   string
	writer     *sftpWriter
	closed     bool
}

func (m *sftpMover) initialize() (err error) {
	// Create a temporary file on the remote server
	var tempNameBytes [16]byte
	if _, err = rand.Read(tempNameBytes[:]); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	tempName := fmt.Sprintf("tmp_%x", tempNameBytes)
	m.tempPath = path.Join(m.store.config.GetRemotePath(), tempName)
	
	if m.tempFile, err = m.store.sftpClient.Create(m.tempPath); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	// Create the streaming writer with compression/encryption
	writeOptions := env_dir.WriteOptions{
		Config: m.config,
		Writer: m.tempFile,
	}
	
	if m.writer, err = newSftpWriter(writeOptions); err != nil {
		m.tempFile.Close()
		m.store.sftpClient.Remove(m.tempPath)
		err = errors.Wrap(err)
		return
	}
	
	return
}

func (m *sftpMover) Write(p []byte) (n int, err error) {
	if m.writer == nil {
		err = errors.ErrorWithStackf("writer not initialized")
		return
	}
	return m.writer.Write(p)
}

func (m *sftpMover) ReadFrom(r io.Reader) (n int64, err error) {
	if m.writer == nil {
		err = errors.ErrorWithStackf("writer not initialized")
		return
	}
	return m.writer.ReadFrom(r)
}

func (m *sftpMover) Close() (err error) {
	if m.closed {
		return nil
	}
	m.closed = true
	
	// Ensure cleanup happens
	defer func() {
		if m.tempFile != nil {
			m.tempFile.Close()
		}
		if m.tempPath != "" {
			m.store.sftpClient.Remove(m.tempPath)
		}
	}()
	
	if m.writer == nil {
		// No data was written
		return nil
	}
	
	// Close the writer to finalize compression/encryption and SHA calculation
	if err = m.writer.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	// Close the temp file
	if err = m.tempFile.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	// Get the calculated SHA and determine final path
	sh := m.writer.GetShaLike()
	finalPath := m.store.remotePathForSha(sh)
	
	// Ensure the target directory exists (Git-like bucketing)
	finalDir := path.Dir(finalPath)
	if err = m.store.sftpClient.MkdirAll(finalDir); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	// Atomically move temp file to final location
	if err = m.store.sftpClient.Rename(m.tempPath, finalPath); err != nil {
		// Check if file already exists
		if _, statErr := m.store.sftpClient.Stat(finalPath); statErr == nil {
			// File already exists, this is OK - just remove temp file
			m.store.sftpClient.Remove(m.tempPath)
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}
	
	// Clear temp path so cleanup doesn't try to remove it
	m.tempPath = ""
	
	return
}

func (m *sftpMover) GetShaLike() interfaces.Sha {
	if m.writer == nil {
		// Return empty SHA if no data written
		return &sha.Sha{}
	}
	return m.writer.GetShaLike()
}

// sftpWriter implements the streaming writer with compression/encryption
type sftpWriter struct {
	hash            hash.Hash
	tee             io.Writer
	wCompress, wAge io.WriteCloser
	wBuf            *bufio.Writer
}

func newSftpWriter(writeOptions env_dir.WriteOptions) (w *sftpWriter, err error) {
	w = &sftpWriter{}

	w.wBuf = bufio.NewWriter(writeOptions.Writer)

	if w.wAge, err = writeOptions.GetBlobEncryption().WrapWriter(w.wBuf); err != nil {
		err = errors.Wrap(err)
		return
	}

	w.hash = sha256.New()

	if w.wCompress, err = writeOptions.GetBlobCompression().WrapWriter(w.wAge); err != nil {
		err = errors.Wrap(err)
		return
	}

	w.tee = io.MultiWriter(w.hash, w.wCompress)

	return
}

func (w *sftpWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if n, err = io.Copy(w.tee, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (w *sftpWriter) Write(p []byte) (n int, err error) {
	return w.tee.Write(p)
}

func (w *sftpWriter) Close() (err error) {
	if w.wCompress != nil {
		if err = w.wCompress.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if w.wAge != nil {
		if err = w.wAge.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if w.wBuf != nil {
		if err = w.wBuf.Flush(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (w *sftpWriter) GetShaLike() interfaces.Sha {
	return sha.FromHash(w.hash)
}

// sftpStreamingReader handles decompression/decryption while reading from SFTP
type sftpStreamingReader struct {
	file   *sftp.File
	config env_dir.Config
}

func (sr *sftpStreamingReader) createReader() (reader interfaces.ShaReadCloser, err error) {
	// Create streaming reader with decompression/decryption
	sftpReader := &sftpReader{
		file:   sr.file,
		config: sr.config,
	}

	if err = sftpReader.initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	reader = sftpReader
	return
}

// sftpReader implements streaming decompression/decryption for SFTP
type sftpReader struct {
	file       *sftp.File
	config     env_dir.Config
	hash       hash.Hash
	decrypter  io.Reader
	expander   io.ReadCloser
	tee        io.Reader
}

func (r *sftpReader) initialize() (err error) {
	// Set up decryption
	if r.decrypter, err = r.config.GetBlobEncryption().WrapReader(r.file); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Set up decompression
	if r.expander, err = r.config.GetBlobCompression().WrapReader(r.decrypter); err != nil {
		// Try without compression if it fails
		if _, err = r.file.Seek(0, io.SeekStart); err != nil {
			err = errors.Wrap(err)
			return
		}

		if r.expander, err = compression_type.CompressionTypeNone.WrapReader(r.file); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// Set up SHA calculation
	r.hash = sha256.New()
	r.tee = io.TeeReader(r.expander, r.hash)

	return
}

func (r *sftpReader) Read(p []byte) (n int, err error) {
	return r.tee.Read(p)
}

func (r *sftpReader) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, r.tee)
}

func (r *sftpReader) Seek(offset int64, whence int) (actual int64, err error) {
	seeker, ok := r.decrypter.(io.Seeker)

	if !ok {
		err = errors.ErrorWithStackf("seeking not supported")
		return
	}

	return seeker.Seek(offset, whence)
}

func (r *sftpReader) Close() error {
	err1 := r.expander.Close()
	err2 := r.file.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

func (r *sftpReader) GetShaLike() interfaces.Sha {
	return sha.FromHash(r.hash)
}