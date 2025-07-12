package blob_stores

import (
	"bytes"
	"fmt"
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
	tempFS     env_dir.TemporaryFS
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

func makeSftpStore(
	config sftpConfig,
	tempFS env_dir.TemporaryFS,
) (store *sftpBlobStore, err error) {
	store = &sftpBlobStore{
		config: config,
		tempFS: tempFS,
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
	// For SFTP, we'll use a temporary local file approach
	var tempFile *os.File
	var tempPath string
	
	if tempFile, err = store.tempFS.FileTempWithTemplate("sftp-blob-*"); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	tempPath = tempFile.Name()
	tempFile.Close()
	
	mover := &sftpMover{
		store:    store,
		tempPath: tempPath,
		writer:   sha.MakeWriter(nil),
	}
	
	// Create a file writer that will write to temp file
	if mover.tempWriter, err = os.Create(tempPath); err != nil {
		err = errors.Wrap(err)
		return
	}
	
	// Set up multi-writer to calculate SHA while writing
	mover.writer = sha.MakeWriter(io.MultiWriter(mover.tempWriter, mover.writer))
	
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
	
	// For reading, we'll download to a temporary file first
	var tempFile *os.File
	if tempFile, err = store.tempFS.FileTempWithTemplate("sftp-read-*"); err != nil {
		err = errors.Wrap(err)
		return
	}
	defer tempFile.Close()
	
	// Download the file
	var remoteFile *sftp.File
	if remoteFile, err = store.sftpClient.Open(remotePath); err != nil {
		os.Remove(tempFile.Name())
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
	defer remoteFile.Close()
	
	// Copy to temp file
	if _, err = io.Copy(tempFile, remoteFile); err != nil {
		os.Remove(tempFile.Name())
		err = errors.Wrap(err)
		return
	}
	
	// Reopen for reading with env_dir to handle compression/encryption
	options := env_dir.FileReadOptions{
		Config: store.makeEnvDirConfig(),
		Path:   tempFile.Name(),
	}
	
	if r, err = env_dir.NewFileReader(options); err != nil {
		os.Remove(tempFile.Name())
		err = errors.Wrap(err)
		return
	}
	
	// Wrap to clean up temp file on close
	r = &cleanupReader{
		ShaReadCloser: r,
		cleanupFunc: func() {
			os.Remove(tempFile.Name())
		},
	}
	
	return
}

// sftpMover implements interfaces.Mover and interfaces.ShaWriteCloser
type sftpMover struct {
	store      *sftpBlobStore
	tempPath   string
	tempWriter io.WriteCloser
	writer     interfaces.ShaWriteCloser
	closed     bool
}

func (m *sftpMover) Write(p []byte) (n int, err error) {
	return m.writer.Write(p)
}

func (m *sftpMover) ReadFrom(r io.Reader) (n int64, err error) {
	return m.writer.ReadFrom(r)
}

func (m *sftpMover) Close() (err error) {
	if m.closed {
		return nil
	}
	m.closed = true
	
	// Close the temporary writer
	if err = m.tempWriter.Close(); err != nil {
		os.Remove(m.tempPath)
		err = errors.Wrap(err)
		return
	}
	
	// Get the calculated SHA
	remotePath := m.store.remotePathForSha(m.writer.GetShaLike())
	
	// Ensure remote directory exists
	remoteDir := path.Dir(remotePath)
	if err = m.store.sftpClient.MkdirAll(remoteDir); err != nil {
		os.Remove(m.tempPath)
		err = errors.Wrap(err)
		return
	}
	
	// Upload the file
	var srcFile *os.File
	if srcFile, err = os.Open(m.tempPath); err != nil {
		os.Remove(m.tempPath)
		err = errors.Wrap(err)
		return
	}
	defer srcFile.Close()
	
	var dstFile *sftp.File
	if dstFile, err = m.store.sftpClient.Create(remotePath); err != nil {
		os.Remove(m.tempPath)
		err = errors.Wrap(err)
		return
	}
	defer dstFile.Close()
	
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		os.Remove(m.tempPath)
		m.store.sftpClient.Remove(remotePath)
		err = errors.Wrap(err)
		return
	}
	
	// Clean up temp file
	os.Remove(m.tempPath)
	
	return
}

func (m *sftpMover) GetShaLike() interfaces.Sha {
	return m.writer.GetShaLike()
}

// cleanupReader wraps a ReadCloser to perform cleanup on Close
type cleanupReader struct {
	interfaces.ShaReadCloser
	cleanupFunc func()
}

func (r *cleanupReader) Close() error {
	err := r.ShaReadCloser.Close()
	r.cleanupFunc()
	return err
}