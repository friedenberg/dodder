package sku_json_fmt

import (
	"bytes"
	"io"
	"net/url"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type TomlBookmark struct {
	Url string `toml:"url"`
}

func TomlBookmarkUrl(
	object *sku.Transacted,
	envRepo env_repo.Env,
) (ur *url.URL, err error) {
	var reader interfaces.ReadCloseBlobIdGetter

	if reader, err = envRepo.GetDefaultBlobStore().BlobReader(object.GetBlobId()); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, reader)

	var buffer bytes.Buffer

	if _, err = io.Copy(&buffer, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	var tb TomlBookmark

	if err = toml.Unmarshal(buffer.Bytes(), &tb); err != nil {
		err = errors.Wrapf(err, "%q", buffer.String())
		return
	}

	if ur, err = url.Parse(tb.Url); err != nil {
		err = errors.Wrapf(err, "%q", tb.Url)
		return
	}

	return
}
