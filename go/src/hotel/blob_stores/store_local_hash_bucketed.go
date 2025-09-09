package blob_stores

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type localHashBucketed struct {
	config blob_store_configs.ConfigLocalHashBucketed

	multiHash       bool
	defaultHashType markl.HashType
	buckets         []int

	basePath string
	tempFS   env_dir.TemporaryFS
}

var _ interfaces.BlobStore = localHashBucketed{}

func makeLocalHashBucketed(
	ctx interfaces.ActiveContext,
	basePath string,
	config blob_store_configs.ConfigLocalHashBucketed,
	tempFS env_dir.TemporaryFS,
) (store localHashBucketed, err error) {
	// TODO read default hash type from config
	if store.defaultHashType, err = markl.GetHashTypeOrError(
		config.GetDefaultHashTypeId(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	store.multiHash = config.SupportsMultiHash()
	store.buckets = config.GetHashBuckets()
	store.config = config
	store.basePath = basePath
	store.tempFS = tempFS

	return
}

func (blobStore localHashBucketed) GetBlobStoreConfig() blob_store_configs.Config {
	return blobStore.config
}

func (blobStore localHashBucketed) GetBlobStoreDescription() string {
	return fmt.Sprintf("TODO: local-git-like")
}

func (blobStore localHashBucketed) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	return blobStore.config
}

func (blobStore localHashBucketed) GetDefaultHashType() interfaces.HashType {
	return blobStore.defaultHashType
}

func (blobStore localHashBucketed) makeEnvDirConfig() env_dir.Config {
	return env_dir.MakeConfig(
		blobStore.defaultHashType,
		env_dir.MakeHashBucketPathJoinFunc(blobStore.buckets),
		blobStore.config.GetBlobCompression(),
		blobStore.config.GetBlobEncryption(),
	)
}

func (blobStore localHashBucketed) HasBlob(
	merkleId interfaces.MarklId,
) (ok bool) {
	if merkleId.IsNull() {
		ok = true
		return
	}

	path := env_dir.MakeHashBucketPathFromMerkleId(
		merkleId,
		blobStore.buckets,
		blobStore.multiHash,
		blobStore.basePath,
	)

	ok = files.Exists(path)

	return
}

func (blobStore localHashBucketed) AllBlobs() interfaces.SeqError[interfaces.MarklId] {
	if blobStore.multiHash {
		return localAllBlobsMultihash(blobStore.basePath)
	} else {
		return localAllBlobs(blobStore.basePath, blobStore.defaultHashType)
	}
}

func (blobStore localHashBucketed) MakeBlobReader(
	digest interfaces.MarklId,
) (readCloser interfaces.BlobReader, err error) {
	if digest.IsNull() {
		readCloser = markl_io.MakeNopReadCloser(
			blobStore.defaultHashType.Get(),
			io.NopCloser(bytes.NewReader(nil)),
		)
		return
	}

	if readCloser, err = blobStore.blobReaderFrom(
		digest,
		blobStore.basePath,
	); err != nil {
		if !env_dir.IsErrBlobMissing(err) {
			err = errors.Wrap(err)
		}

		return
	}

	return
}

func (blobStore localHashBucketed) MakeBlobWriter(
	marklHashType interfaces.HashType,
) (blobWriter interfaces.BlobWriter, err error) {
	if blobWriter, err = blobStore.blobWriterTo(
		blobStore.basePath,
		marklHashType,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore localHashBucketed) blobWriterTo(
	path string,
	marklHashType interfaces.HashType,
) (mover interfaces.BlobWriter, err error) {
	if blobStore.multiHash {
		path = filepath.Join(path, blobStore.defaultHashType.GetMarklTypeId())
	}

	if mover, err = env_dir.NewMover(
		blobStore.makeEnvDirConfig(),
		env_dir.MoveOptions{
			FinalPathOrDir:              path,
			GenerateFinalPathFromDigest: true,
			TemporaryFS:                 blobStore.tempFS,
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore localHashBucketed) blobReaderFrom(
	digest interfaces.MarklId,
	path string,
) (readCloser interfaces.BlobReader, err error) {
	if digest.IsNull() {
		readCloser = markl_io.MakeNopReadCloser(
			blobStore.defaultHashType.Get(),
			io.NopCloser(bytes.NewReader(nil)),
		)
		return
	}

	marklType := digest.GetMarklType()

	if marklType == nil {
		err = errors.Errorf("empty markl type")
		return
	}

	if marklType.GetMarklTypeId() == "" {
		err = errors.Errorf("empty markl type id")
		return
	}

	path = env_dir.MakeHashBucketPathFromMerkleId(
		digest,
		blobStore.buckets,
		blobStore.multiHash,
		path,
	)

	if readCloser, err = env_dir.NewFileReader(
		blobStore.makeEnvDirConfig(),
		path,
	); err != nil {
		if errors.IsNotExist(err) {
			err = env_dir.ErrBlobMissing{
				BlobId: markl.Clone(digest),
				Path:   path,
			}
		} else {
			err = errors.Wrapf(
				err,
				"Path: %q, Compression: %q",
				path,
				blobStore.config.GetBlobCompression(),
			)
		}

		return
	}

	return
}
