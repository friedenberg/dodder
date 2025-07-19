package sku_fmt

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Json struct {
	BlobSha     string        `json:"blob-sha"`
	BlobString  string        `json:"blob-string"`
	Date        string        `json:"date"`
	Description string        `json:"description"`
	Dormant     bool          `json:"dormant"`
	ObjectId    string        `json:"object-id"`
	RepoPubkey  blech32.Value `json:"repo-pub_key"`
	RepoSig     blech32.Value `json:"repo-sig"`
	Sha         string        `json:"sha"`
	Tags        []string      `json:"tags"`
	Tai         string        `json:"tai"`
	Type        string        `json:"type"`
}

func (json *Json) FromStringAndMetadata(
	objectId string,
	metadata *object_metadata.Metadata,
	envRepo env_repo.Env,
) (err error) {
	var readCloser sha.ReadCloser

	if readCloser, err = envRepo.GetDefaultBlobStore().BlobReader(&metadata.Blob); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	var blobStringBuilder strings.Builder

	if _, err = io.Copy(&blobStringBuilder, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	json.BlobSha = metadata.Blob.String()
	json.BlobString = blobStringBuilder.String()
	json.Date = metadata.Tai.Format(string_format_writer.StringFormatDateTime)
	json.Description = metadata.Description.String()
	json.Dormant = metadata.Cache.Dormant.Bool()
	json.ObjectId = objectId
	json.RepoPubkey = metadata.GetRepoPubkeyValue()
	json.RepoSig = metadata.GetRepoSigValue()
	json.Sha = metadata.SelfMetadataWithoutTai.String()
	json.Tags = quiter.Strings(metadata.GetTags())
	json.Tai = metadata.Tai.String()
	json.Type = metadata.Type.String()
	// TODO add support for "preview"

	return
}

func (json *Json) FromTransacted(
	object *sku.Transacted,
	envRepo env_repo.Env,
) (err error) {
	return json.FromStringAndMetadata(
		object.ObjectId.String(),
		object.GetMetadata(),
		envRepo,
	)
}

func (json *Json) ToTransacted(
	object *sku.Transacted,
	envRepo env_repo.Env,
) (err error) {
	var writeCloser sha.WriteCloser

	if writeCloser, err = envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if _, err = io.Copy(writeCloser, strings.NewReader(json.BlobString)); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO-P1 support states of blob vs blob sha
	object.SetBlobSha(writeCloser.GetShaLike())

	if err = object.ObjectId.Set(json.ObjectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = object.Metadata.Type.Set(json.Type); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = object.Metadata.Description.Set(json.Description); err != nil {
		err = errors.Wrap(err)
		return
	}

	var es ids.TagSet

	if es, err = ids.MakeTagSetStrings(json.Tags...); err != nil {
		err = errors.Wrap(err)
		return
	}

	object.Metadata.SetTags(es)
	object.Metadata.GenerateExpandedTags()

	object.Metadata.RepoPubkey = json.RepoPubkey.Data
	object.Metadata.RepoSig = json.RepoSig.Data

	return
}
