package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config"
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

	return repo
}

func (local *Repo) initDefaultTypeAndConfig(
	bigBang env_repo.BigBang,
) (err error) {
	// TODO determine if this lock/unlock is necessary
	if err = local.Lock(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.Deferred(&err, local.Unlock)

	var defaultTypeObjectId ids.Type

	if defaultTypeObjectId, err = local.initDefaultTypeIfNecessaryAfterLock(
		bigBang,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = local.initDefaultConfigIfNecessaryAfterLock(
		bigBang,
		local.GetEnvRepo().GetDefaultBlobStore().GetName(),
		defaultTypeObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (local *Repo) initDefaultTypeIfNecessaryAfterLock(
	bigBang env_repo.BigBang,
) (defaultTypeObjectId ids.Type, err error) {
	if bigBang.ExcludeDefaultType {
		return defaultTypeObjectId, err
	}

	defaultTypeObjectId = ids.MustType("md")
	defaultTypeBlob := type_blobs.Default()

	var objectId ids.ObjectId

	if err = objectId.SetWithIdLike(defaultTypeObjectId); err != nil {
		err = errors.Wrap(err)
		return defaultTypeObjectId, err
	}

	var sh interfaces.MarklId

	// TODO remove and replace with two-step process
	if sh, _, err = local.GetStore().GetTypedBlobStore().GetTypeV1().SaveBlobText(
		&defaultTypeBlob,
	); err != nil {
		err = errors.Wrap(err)
		return defaultTypeObjectId, err
	}

	object := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(object)

	if err = object.ObjectId.SetWithIdLike(defaultTypeObjectId); err != nil {
		err = errors.Wrap(err)
		return defaultTypeObjectId, err
	}

	object.Metadata.GetBlobDigestMutable().ResetWithMarklId(sh)
	object.GetMetadata().Type = ids.DefaultOrPanic(genres.Type)

	if err = local.GetStore().CreateOrUpdateDefaultProto(
		object,
		sku.GetStoreOptionsCreate(),
	); err != nil {
		err = errors.Wrap(err)
		return defaultTypeObjectId, err
	}

	return defaultTypeObjectId, err
}

func (local *Repo) initDefaultConfigIfNecessaryAfterLock(
	bigBang env_repo.BigBang,
	defaultBlobStoreName string,
	defaultTypeObjectId ids.Type,
) (err error) {
	if bigBang.ExcludeDefaultConfig {
		return err
	}

	var blobId interfaces.MarklId
	var typedBlob repo_config.TypedBlob

	if blobId, typedBlob, err = writeDefaultMutableConfig(
		local,
		defaultBlobStoreName,
		defaultTypeObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	newConfig := sku.GetTransactedPool().Get()

	if err = newConfig.ObjectId.SetWithIdLike(ids.Config{}); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = newConfig.SetBlobDigest(blobId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	newConfig.Metadata.Type.ResetWith(typedBlob.Type)

	if err = local.GetStore().CreateOrUpdateDefaultProto(
		newConfig,
		sku.GetStoreOptionsCreate(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func writeDefaultMutableConfig(
	repo *Repo,
	defaultBlobStoreName string,
	defaultType ids.Type,
) (blobId interfaces.MarklId, typedBlob repo_config.TypedBlob, err error) {
	typedBlob = repo_config.DefaultOverlay(defaultBlobStoreName, defaultType)

	coder := repo.GetStore().GetConfigBlobCoder()

	var writeCloser interfaces.BlobWriter

	if writeCloser, err = repo.GetEnvRepo().GetDefaultBlobStore().MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return blobId, typedBlob, err
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if _, err = coder.EncodeTo(
		&typedBlob,
		writeCloser,
	); err != nil {
		err = errors.Wrap(err)
		return blobId, typedBlob, err
	}

	blobId = writeCloser.GetMarklId()

	return blobId, typedBlob, err
}
