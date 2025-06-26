package sku_fmt

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Json struct {
	BlobString  string   `json:"blob-string"`
	BlobSha     string   `json:"blob-sha"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	ObjectId    string   `json:"object-id"`
	Sha         string   `json:"sha"`
	Type        string   `json:"type"`
	Tai         string   `json:"tai"`
	Date        string   `json:"date"`
}

func (j *Json) FromStringAndMetadata(
	k string,
	m *object_metadata.Metadata,
	s env_repo.Env,
) (err error) {
	var r sha.ReadCloser

	if r, err = s.BlobReader(&m.Blob); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, r)

	var out strings.Builder

	if _, err = io.Copy(&out, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	j.BlobString = out.String()
	j.BlobSha = m.Blob.String()
	j.Description = m.Description.String()
	j.Tags = quiter.Strings(m.GetTags())
	j.ObjectId = k
	j.Sha = m.SelfMetadataWithoutTai.String()
	j.Tai = m.Tai.String()
	j.Date = m.Tai.Format(string_format_writer.StringFormatDateTime)
	j.Type = m.Type.String()
	// TODO add support for "preview"

	return
}

func (j *Json) FromTransacted(
	sk *sku.Transacted,
	s env_repo.Env,
) (err error) {
	return j.FromStringAndMetadata(sk.ObjectId.String(), sk.GetMetadata(), s)
}

func (j *Json) ToTransacted(sk *sku.Transacted, s env_repo.Env) (err error) {
	var w sha.WriteCloser

	if w, err = s.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, w)

	if _, err = io.Copy(w, strings.NewReader(j.BlobString)); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO-P1 support states of blob vs blob sha
	sk.SetBlobSha(w.GetShaLike())

	if err = sk.ObjectId.Set(j.ObjectId); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = sk.Metadata.Type.Set(j.Type); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = sk.Metadata.Description.Set(j.Description); err != nil {
		err = errors.Wrap(err)
		return
	}

	var es ids.TagSet

	if es, err = ids.MakeTagSetStrings(j.Tags...); err != nil {
		err = errors.Wrap(err)
		return
	}

	sk.Metadata.SetTags(es)
	sk.Metadata.GenerateExpandedTags()

	return
}
