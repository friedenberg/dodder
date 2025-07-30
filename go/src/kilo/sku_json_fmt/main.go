package sku_json_fmt

import (
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
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Transacted struct {
	BlobId      string        `json:"blob-id"`
	BlobString  string        `json:"blob-string,omitempty"`
	Date        string        `json:"date"`
	Description string        `json:"description"`
	ObjectId    string        `json:"object-id"`
	RepoPubkey  blech32.Value `json:"repo-pub_key"`
	RepoSig     blech32.Value `json:"repo-sig"`
	Sha         string        `json:"sha"`
	Tags        []string      `json:"tags"`
	Tai         string        `json:"tai"`
	Type        string        `json:"type"`
}

func (json *Transacted) FromStringAndMetadata(
	objectId string,
	metadata *object_metadata.Metadata,
	blobStore interfaces.BlobStore,
) (err error) {
	if blobStore != nil {
		var readCloser interfaces.ReadCloseBlobIdGetter

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

	json.BlobId = metadata.Blob.String()
	json.Date = metadata.Tai.Format(string_format_writer.StringFormatDateTime)
	json.Description = metadata.Description.String()
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

func (json *Transacted) FromTransacted(
	object *sku.Transacted,
	blobStore interfaces.BlobStore,
) (err error) {
	return json.FromStringAndMetadata(
		object.ObjectId.String(),
		object.GetMetadata(),
		blobStore,
	)
}

func (json *Transacted) ToTransacted(
	object *sku.Transacted,
	blobStore interfaces.BlobStore,
) (err error) {
	if blobStore != nil {
		var writeCloser interfaces.WriteCloseBlobIdGetter

		if writeCloser, err = blobStore.BlobWriter(); err != nil {
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

		// TODO just compare blob digests
		// TODO-P1 support states of blob vs blob sha
		object.SetBlobId(writeCloser.GetBlobId())
	}

	// Set BlobId from JSON even if not writing to blob store
	if json.BlobId != "" && blobStore == nil {
		if err = object.Metadata.Blob.Set(json.BlobId); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = object.ObjectId.Set(json.ObjectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO enforce non-empty types
	if json.Type != "" {
		if err = object.Metadata.Type.Set(json.Type); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = object.Metadata.Description.Set(json.Description); err != nil {
		err = errors.Wrap(err)
		return
	}

	var tagSet ids.TagSet

	if tagSet, err = ids.MakeTagSetStrings(json.Tags...); err != nil {
		err = errors.Wrap(err)
		return
	}

	object.Metadata.SetTags(tagSet)
	object.Metadata.GenerateExpandedTags()

	object.Metadata.RepoPubkey = json.RepoPubkey.Data
	object.Metadata.RepoSig = json.RepoSig.Data

	// Set Tai from either Date or Tai field
	if json.Tai != "" {
		if err = object.Metadata.Tai.Set(json.Tai); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else if json.Date != "" {
		if err = object.Metadata.Tai.SetFromRFC3339(json.Date); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// Set SelfMetadataWithoutTai SHA
	if json.Sha != "" {
		if err = object.Metadata.SelfMetadataWithoutTai.Set(json.Sha); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
