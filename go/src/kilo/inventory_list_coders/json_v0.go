package inventory_list_coders

import (
	"bufio"
	"encoding/json"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_fmt"
)

type JSONV0 struct {
	ImmutableConfigPrivate genesis_configs.ConfigPrivate
}

func (coder JSONV0) GetType() ids.Type {
	return ids.MustType(ids.TypeInventoryListV2)
}

func (coder JSONV0) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if object.Metadata.GetDigest().IsNull() {
		err = errors.ErrorWithStackf("empty digest: %q", sku.String(object))
		return
	}

	if object.Metadata.RepoSig.IsEmpty() {
		err = errors.ErrorWithStackf("no repo signature")
		return
	}

	var objectJson sku_fmt.JSON

	if err = objectJson.FromTransacted(object, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	encoder := json.NewEncoder(bufferedWriter)

	if err = encoder.Encode(objectJson); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coder JSONV0) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var objectJson sku_fmt.JSON

	decoder := json.NewDecoder(bufferedReader)

	if err = decoder.Decode(&objectJson); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = objectJson.ToTransacted(object, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	if object.GetType().String() == ids.TypeInventoryListV2 {
		sh := sha.MustWithDigester(object.GetTai())
		defer digests.PutBlobId(sh)

		if len(object.Metadata.RepoPubkey) == 0 {
			err = errors.ErrorWithStackf(
				"RepoPubkey missing for %s. Fields: %#v",
				sku.String(object),
				object.Metadata.Fields,
			)
			return
		}

		if object.Metadata.RepoSig.IsEmpty() {
			err = errors.ErrorWithStackf(
				"signature missing for %s. Fields: %#v",
				sku.String(object),
				object.Metadata.Fields,
			)
			return
		}

		if err = repo_signing.VerifySignature(
			object.Metadata.RepoPubkey,
			sh.GetBytes(),
			object.Metadata.RepoSig,
		); err != nil {
			err = errors.Wrapf(
				err,
				"Sku: %s, Tags %s",
				sku.String(object),
				quiter.StringCommaSeparated(object.Metadata.GetTags()),
			)
			return
		}
	} else {
		// TODO determine how to handle this
	}

	if err = object.CalculateObjectDigests(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
