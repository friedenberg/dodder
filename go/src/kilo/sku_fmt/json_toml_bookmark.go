package sku_fmt

import (
	"net/url"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type JsonWithUrl struct {
	Json
	TomlBookmark
}

func MakeJsonTomlBookmark(
	object *sku.Transacted,
	envRepo env_repo.Env,
	tabs []any,
) (json JsonWithUrl, err error) {
	if err = json.FromTransacted(object, envRepo.GetDefaultBlobStore()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = toml.Unmarshal([]byte(json.BlobString), &json.TomlBookmark); err != nil {
		err = errors.Wrapf(err, "%q", json.BlobString)
		return
	}

	if _, err = url.Parse(json.Url); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, tabRaw := range tabs {
		tab := tabRaw.(map[string]interface{})

		if _, err = url.Parse(tab["url"].(string)); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
