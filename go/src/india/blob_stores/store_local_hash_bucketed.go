package blob_stores

import (
	"bytes"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/markl_io"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
)

type localHashBucketed struct {
	config blob_store_configs.ConfigLocalHashBucketed

	multiHash         bool
	defaultHashFormat markl.FormatHash
	buckets           []int

	basePath string
	tempFS   env_dir.TemporaryFS
}

var _ interfaces.BlobStore = localHashBucketed{}

func makeLocalHashBucketed(
	envDir env_dir.Env,
	basePath string,
	config blob_store_configs.ConfigLocalHashBucketed,
) (store localHashBucketed, err error) {
	store.config = config

	store.multiHash = config.SupportsMultiHash()
	if store.defaultHashFormat, err = markl.GetFormatHashOrError(
		config.GetDefaultHashTypeId(),
	); err != nil {
		err = errors.Wrap(err)
		return store, err
	}
	store.buckets = config.GetHashBuckets()

	store.basePath = basePath
	store.tempFS = envDir.GetTempLocal()

	return store, err
}

func (blobStore localHashBucketed) GetBlobStoreConfig() blob_store_configs.Config {
	return blobStore.config
}

func (blobStore localHashBucketed) GetBlobStoreDescription() string {
	return "local hash bucketed"
}

func (blobStore localHashBucketed) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	return blobStore.config
}

func (blobStore localHashBucketed) GetDefaultHashType() interfaces.FormatHash {
	return blobStore.defaultHashFormat
}

func (blobStore localHashBucketed) makeEnvDirConfig(
	hashFormat interfaces.FormatHash,
) env_dir.Config {
	if hashFormat == nil {
		hashFormat = blobStore.defaultHashFormat
	}

	return env_dir.MakeConfig(
		hashFormat,
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
		return ok
	}

	path := env_dir.MakeHashBucketPathFromMerkleId(
		merkleId,
		blobStore.buckets,
		blobStore.multiHash,
		blobStore.basePath,
	)

	ok = files.Exists(path)

	return ok
}

func (blobStore localHashBucketed) AllBlobs() interfaces.SeqError[interfaces.MarklId] {
	if blobStore.multiHash {
		return localAllBlobsMultihash(blobStore.basePath)
	} else {
		return localAllBlobs(blobStore.basePath, blobStore.defaultHashFormat)
	}
}

func (blobStore localHashBucketed) MakeBlobReader(
	digest interfaces.MarklId,
) (readCloser interfaces.BlobReader, err error) {
	if digest.IsNull() {
		readCloser = markl_io.MakeNopReadCloser(
			blobStore.defaultHashFormat.Get(),
			ohio.NopCloser(bytes.NewReader(nil)),
		)
		return readCloser, err
	}

	if readCloser, err = blobStore.blobReaderFrom(
		digest,
		blobStore.basePath,
	); err != nil {
		if !env_dir.IsErrBlobMissing(err) {
			err = errors.Wrap(err)
		}

		return readCloser, err
	}

	return readCloser, err
}

func (blobStore localHashBucketed) MakeBlobWriter(
	marklHashType interfaces.FormatHash,
) (blobWriter interfaces.BlobWriter, err error) {
	if blobWriter, err = blobStore.blobWriterTo(
		blobStore.basePath,
		marklHashType,
	); err != nil {
		err = errors.Wrap(err)
		return blobWriter, err
	}

	return blobWriter, err
}

func (blobStore localHashBucketed) blobWriterTo(
	path string,
	hashFormat interfaces.FormatHash,
) (mover interfaces.BlobWriter, err error) {
	if hashFormat == nil {
		hashFormat = blobStore.defaultHashFormat
	}

	if blobStore.multiHash {
		path = filepath.Join(
			path,
			hashFormat.GetMarklFormatId(),
		)
	}

	if mover, err = env_dir.NewMover(
		blobStore.makeEnvDirConfig(hashFormat),
		env_dir.MoveOptions{
			FinalPathOrDir:              path,
			GenerateFinalPathFromDigest: true,
			TemporaryFS:                 blobStore.tempFS,
		},
	); err != nil {
		err = errors.Wrap(err)
		return mover, err
	}

	return mover, err
}

func (blobStore localHashBucketed) blobReaderFrom(
	digest interfaces.MarklId,
	basePath string,
) (readCloser interfaces.BlobReader, err error) {
	if digest.IsNull() {
		readCloser = markl_io.MakeNopReadCloser(
			blobStore.defaultHashFormat.Get(),
			ohio.NopCloser(bytes.NewReader(nil)),
		)
		return readCloser, err
	}

	marklType := digest.GetMarklFormat()

	if marklType == nil {
		err = errors.Errorf("empty markl type")
		return readCloser, err
	}

	if marklType.GetMarklFormatId() == "" {
		err = errors.Errorf("empty markl type id")
		return readCloser, err
	}

	basePath = env_dir.MakeHashBucketPathFromMerkleId(
		digest,
		blobStore.buckets,
		blobStore.multiHash,
		basePath,
	)

	if readCloser, err = env_dir.NewFileReaderOrErrNotExist(
		blobStore.makeEnvDirConfig(nil),
		basePath,
	); err != nil {
		if errors.IsNotExist(err) {
			err = env_dir.ErrBlobMissing{
				BlobId: markl.Clone(digest),
				Path:   basePath,
			}
		} else {
			err = errors.Wrapf(
				err,
				"Path: %q, Compression: %q",
				basePath,
				blobStore.config.GetBlobCompression(),
			)
		}

		return readCloser, err
	}

	return readCloser, err
}
