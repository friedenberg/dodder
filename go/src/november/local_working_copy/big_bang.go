package local_working_copy

import (
	"encoding/gob"
	"io"
	"os"
	"path"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/golf/config_mutable_blobs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func Genesis(
	bb env_repo.BigBang,
	repoLayout env_repo.Env,
) (repo *Repo) {
	repo = MakeWithLayout(OptionsEmpty, repoLayout)

	if err := repo.dormantIndex.Flush(
		repo.GetEnvRepo(),
		repo.PrinterHeader(),
		repo.config.GetCLIConfig().IsDryRun(),
	); err != nil {
		repo.CancelWithError(err)
	}

	repo.Must(repo.Reset)
	repo.Must(repo.envRepo.ResetCache)

	if err := repo.initDefaultTypeAndConfig(bb); err != nil {
		repo.CancelWithError(err)
	}

	repo.Must(repo.Lock)
	repo.Must(repo.GetStore().ResetIndexes)
	repo.Must(repo.Unlock)

	return
}

func (local *Repo) initDefaultTypeAndConfig(bb env_repo.BigBang) (err error) {
	if err = local.Lock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.Deferred(&err, local.Unlock)

	var defaultTypeObjectId ids.Type

	if defaultTypeObjectId, err = local.initDefaultTypeIfNecessaryAfterLock(
		bb,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = local.initDefaultConfigIfNecessaryAfterLock(
		bb,
		defaultTypeObjectId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (local *Repo) initDefaultTypeIfNecessaryAfterLock(
	bb env_repo.BigBang,
) (defaultTypeObjectId ids.Type, err error) {
	if bb.ExcludeDefaultType {
		return
	}

	defaultTypeObjectId = ids.MustType("md")
	defaultTypeBlob := type_blobs.Default()

	var k ids.ObjectId

	if err = k.SetWithIdLike(defaultTypeObjectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	var sh interfaces.Sha

	// TODO remove and replace with two-step process
	if sh, _, err = local.GetStore().GetTypedBlobStore().GetTypeV1().SaveBlobText(
		&defaultTypeBlob,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	o := sku.GetTransactedPool().Get()
	defer sku.GetTransactedPool().Put(o)

	if err = o.ObjectId.SetWithIdLike(defaultTypeObjectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	o.Metadata.Blob.ResetWithShaLike(sh)
	o.GetMetadata().Type = builtin_types.DefaultOrPanic(genres.Type)

	if err = local.GetStore().CreateOrUpdateDefaultProto(
		o,
		sku.GetStoreOptionsCreate(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (local *Repo) initDefaultConfigIfNecessaryAfterLock(
	bb env_repo.BigBang,
	defaultTypeObjectId ids.Type,
) (err error) {
	if bb.ExcludeDefaultConfig {
		return
	}

	var sh interfaces.Sha
	var tipe ids.Type

	if sh, tipe, err = writeDefaultMutableConfig(
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

	if err = newConfig.SetBlobSha(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	newConfig.Metadata.Type.ResetWith(tipe)

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
	u *Repo,
	dt ids.Type,
) (sh interfaces.Sha, tipe ids.Type, err error) {
	defaultMutableConfig := config_mutable_blobs.Default(dt)
	tipe = defaultMutableConfig.Type

	f := u.GetStore().GetConfigBlobFormat()

	var aw sha.WriteCloser

	if aw, err = u.GetEnvRepo().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, aw)

	if _, err = f.EncodeTo(defaultMutableConfig.Blob, aw); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh = sha.Make(aw.GetShaLike())

	return
}

func mkdirAll(elements ...string) {
	err := os.MkdirAll(path.Join(elements...), os.ModeDir|0o755)
	errors.PanicIfError(err)
}

func writeFile(p string, contents any) {
	var f *os.File
	var err error

	if f, err = files.CreateExclusiveWriteOnly(p); err != nil {
		if errors.IsExist(err) {
			ui.Err().Printf("%s already exists, not overwriting", p)
			err = nil
		} else {
		}

		return
	}

	defer errors.PanicIfError(err)
	defer errors.DeferredCloser(&err, f)

	if s, ok := contents.(string); ok {
		_, err = io.WriteString(f, s)
	} else {
		enc := gob.NewEncoder(f)
		err = enc.Encode(contents)
	}
}
