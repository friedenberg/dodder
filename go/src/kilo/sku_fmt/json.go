package sku_fmt

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type JsonMCP struct {
	URI         string   `json:"uri,omitempty"`
	RelatedURIs []string `json:"related_uris,omitempty"`
}

type Json struct {
	// TODO rename to blob-digest
	BlobDigest  string        `json:"blob-sha"`
	BlobString  string        `json:"blob-string,omitempty"`
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

	JsonMCP
}

func (json *Json) FromStringAndMetadata(
	objectId string,
	metadata *object_metadata.Metadata,
	blobStore interfaces.LocalBlobStore,
) (err error) {
	if blobStore != nil {
		var readCloser interfaces.ReadCloseDigester

		if readCloser, err = blobStore.BlobReader(&metadata.Blob); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, readCloser)

		var blobStringBuilder strings.Builder

		if _, err = io.Copy(&blobStringBuilder, readCloser); err != nil {
			err = errors.Wrap(err)
			return
		}

		json.BlobString = blobStringBuilder.String()
	}

	json.BlobDigest = metadata.Blob.String()
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

	json.URI = fmt.Sprintf("dodder:///objects/%s", objectId)

	json.RelatedURIs = make([]string, 0, 2+len(json.Tags))

	json.RelatedURIs = append(
		json.RelatedURIs,
		fmt.Sprintf("dodder:///objects/%s", metadata.GetType()),
	)

	for _, tag := range json.Tags {
		json.RelatedURIs = append(
			json.RelatedURIs,
			fmt.Sprintf("dodder:///objects/%s", tag),
		)
	}

	// json.MCPURI = fmt.Sprintf(
	// 	"dodder://objects/%s@%s:%s",
	// 	objectId,
	// 	json.RepoPubkey,
	// 	json.RepoSig,
	// )

	// TODO add support for "preview"

	return
}

// TODO accept blob store instead of env
func (json *Json) FromTransacted(
	object *sku.Transacted,
	blobStore interfaces.LocalBlobStore,
) (err error) {
	return json.FromStringAndMetadata(
		object.ObjectId.String(),
		object.GetMetadata(),
		blobStore,
	)
}

func (json *Json) ToTransacted(
	object *sku.Transacted,
	envRepo env_repo.Env,
) (err error) {
	var writeCloser interfaces.WriteCloseDigester

	if writeCloser, err = envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writeCloser)

	reader, repool := pool.GetStringReader(json.BlobString)
	defer repool()
	if _, err = io.Copy(writeCloser, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO-P1 support states of blob vs blob sha
	object.SetBlobSha(writeCloser.GetDigest())

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
