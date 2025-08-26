package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func Genesis(
	bigBang env_repo.BigBang,
	envRepo env_repo.Env,
) (repo *Repo) {
	repo = MakeWithEnvRepo(OptionsEmpty, envRepo)

	if err := repo.dormantIndex.Flush(
		repo.GetEnvRepo(),
		repo.PrinterHeader(),
		repo.config.GetConfig().IsDryRun(),
	); err != nil {
		repo.Cancel(err)
	}

	repo.Must(errors.MakeFuncContextFromFuncErr(repo.Reset))
	repo.Must(errors.MakeFuncContextFromFuncErr(repo.envRepo.ResetCache))

	if err := repo.initDefaultTypeAndConfig(bigBang); err != nil {
		repo.Cancel(err)
	}

	repo.Must(errors.MakeFuncContextFromFuncErr(repo.Lock))
	repo.Must(errors.MakeFuncContextFromFuncErr(repo.GetStore().ResetIndexes))
	repo.Must(errors.MakeFuncContextFromFuncErr(repo.Unlock))

	return
}

func (local *Repo) initDefaultTypeAndConfig(
	bigBang env_repo.BigBang,
) (err error) {
	// TODO determine if this lock/unlock is necessary
	if err = local.Lock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.Deferred(&err, local.Unlock)

	var defaultTypeObjectId ids.Type

	if defaultTypeObjectId, err = local.initDefaultTypeIfNecessaryAfterLock(
		bigBang,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = local.initDefaultConfigIfNecessaryAfterLock(
		bigBang,
		defaultTypeObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (local *Repo) initDefaultTypeIfNecessaryAfterLock(
	bigBang env_repo.BigBang,
) (defaultTypeObjectId ids.Type, err error) {
	if bigBang.ExcludeDefaultType {
		return
	}

	defaultTypeObjectId = ids.MustType("md")
	defaultTypeBlob := type_blobs.Default()

	var objectId ids.ObjectId

	if err = objectId.SetWithIdLike(defaultTypeObjectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	var sh interfaces.BlobId

	// TODO remove and replace with two-step process
	if sh, _, err = local.GetStore().GetTypedBlobStore().GetTypeV1().SaveBlobText(
		&defaultTypeBlob,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	object := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(object)

	if err = object.ObjectId.SetWithIdLike(defaultTypeObjectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	object.Metadata.GetBlobDigestMutable().ResetWithMerkleId(sh)
	object.GetMetadata().Type = ids.DefaultOrPanic(genres.Type)

	if err = local.GetStore().CreateOrUpdateDefaultProto(
		object,
		sku.GetStoreOptionsCreate(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (local *Repo) initDefaultConfigIfNecessaryAfterLock(
	bigBang env_repo.BigBang,
	defaultTypeObjectId ids.Type,
) (err error) {
	if bigBang.ExcludeDefaultConfig {
		return
	}

	var blobId interfaces.BlobId
	var typedBlob repo_configs.TypedBlob

	if blobId, typedBlob, err = writeDefaultMutableConfig(
		local,
		defaultTypeObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	newConfig := sku.GetTransactedPool().Get()

	if err = newConfig.ObjectId.SetWithIdLike(ids.Config{}); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = newConfig.SetBlobDigest(blobId); err != nil {
		err = errors.Wrap(err)
		return
	}

	newConfig.Metadata.Type.ResetWith(typedBlob.Type)

	if err = local.GetStore().CreateOrUpdateDefaultProto(
		newConfig,
		sku.GetStoreOptionsCreate(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func writeDefaultMutableConfig(
	repo *Repo,
	defaultType ids.Type,
) (blobId interfaces.BlobId, typedBlob repo_configs.TypedBlob, err error) {
	typedBlob = repo_configs.DefaultOverlay(defaultType)

	coder := repo.GetStore().GetConfigBlobCoder()

	var writeCloser interfaces.WriteCloseBlobIdGetter

	if writeCloser, err = repo.GetEnvRepo().GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if _, err = coder.EncodeTo(
		&typedBlob,
		writeCloser,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	blobId = sha.MustWithDigester(writeCloser)

	return
}
