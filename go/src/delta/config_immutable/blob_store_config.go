package config_immutable

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/age"
)

type BlobStoreType string

const (
	BlobStoreTypeLocal  BlobStoreType = "local"
	BlobStoreTypeRemote BlobStoreType = "remote"
	BlobStoreTypeS3     BlobStoreType = "s3"
	BlobStoreTypeGCS    BlobStoreType = "gcs"
)

type BlobStoreConfig interface {
	interfaces.BlobStoreConfigImmutable
	GetBlobStoreType() BlobStoreType
	WriteTo(w io.Writer) (n int64, err error)
	WriteToFile(path string) (err error)
}

type TomlLocalBlobStoreConfigV1 struct {
	Type              BlobStoreType   `toml:"type"`
	AgeEncryption     age.Age         `toml:"age-encryption,omitempty"`
	CompressionType   CompressionType `toml:"compression-type"`
	LockInternalFiles bool            `toml:"lock-internal-files"`
}

const BlobStoreConfigFilename = ".dodder.blob-store.toml"

func (c *TomlLocalBlobStoreConfigV1) GetBlobStoreConfigImmutable() interfaces.BlobStoreConfigImmutable {
	return c
}

func (c *TomlLocalBlobStoreConfigV1) GetBlobEncryption() interfaces.BlobEncryption {
	return &c.AgeEncryption
}

func (c *TomlLocalBlobStoreConfigV1) GetBlobCompression() interfaces.BlobCompression {
	return &c.CompressionType
}

func (c *TomlLocalBlobStoreConfigV1) GetLockInternalFiles() bool {
	return c.LockInternalFiles
}

func (c *TomlLocalBlobStoreConfigV1) GetBlobStoreType() BlobStoreType {
	return c.Type
}

func (c *TomlLocalBlobStoreConfigV1) WriteTo(w io.Writer) (n int64, err error) {
	bufferedWriter := bufio.NewWriter(w)
	defer errors.DeferredFlusher(&err, bufferedWriter)

	encoder := toml.NewEncoder(bufferedWriter)

	if err = encoder.Encode(c); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (c *TomlLocalBlobStoreConfigV1) WriteToFile(path string) (err error) {
	var f *os.File
	if f, err = files.CreateExclusiveWriteOnly(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	if _, err = c.WriteTo(f); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func ReadBlobStoreConfig(path string) (config BlobStoreConfig, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	decoder := toml.NewDecoder(f)

	var tomlConfig TomlLocalBlobStoreConfigV1
	if err = decoder.Decode(&tomlConfig); err != nil {
		err = errors.Wrap(err)
		return
	}

	config = &tomlConfig
	return
}

func ReadBlobStoreConfigFromDir(dir string) (config BlobStoreConfig, err error) {
	path := filepath.Join(dir, BlobStoreConfigFilename)
	return ReadBlobStoreConfig(path)
}

func DefaultLocalBlobStoreConfig() BlobStoreConfig {
	return &TomlLocalBlobStoreConfigV1{
		Type:              BlobStoreTypeLocal,
		CompressionType:   CompressionTypeDefault,
		LockInternalFiles: true,
	}
}

func DefaultBlobStoreConfigForType(blobStoreType BlobStoreType) BlobStoreConfig {
	switch blobStoreType {
	case BlobStoreTypeLocal:
		return DefaultLocalBlobStoreConfig()
	default:
		return DefaultLocalBlobStoreConfig()
	}
}
