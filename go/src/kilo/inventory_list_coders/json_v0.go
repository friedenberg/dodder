package inventory_list_coders

import (
	"bufio"
	"encoding/json"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_json_fmt"
)

type jsonV0 struct {
	ImmutableConfigPrivate genesis_configs.ConfigPrivate
}

func (coder jsonV0) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if object.Metadata.GetDigest().IsNull() {
		err = errors.ErrorWithStackf("empty digest: %q", sku.String(object))
		return
	}

	if object.Metadata.GetContentSig().IsNull() {
		err = errors.ErrorWithStackf("no repo signature")
		return
	}

	var objectJson sku_json_fmt.Transacted

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

func (coder jsonV0) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var objectJson sku_json_fmt.Transacted

	bytess, err := bufferedReader.ReadBytes('\n')
	if err != nil {
		return
	}

	if err = json.Unmarshal(bytess, &objectJson); err != nil {
		err = errors.Wrapf(err, "Line: %q", bytess)
		return
	}

	if err = objectJson.ToTransacted(object, nil); err != nil {
		err = errors.Wrapf(err, "Line: %q", bytess)
		return
	}

	if object.GetType().String() == ids.TypeInventoryListV2 {
		digest := sha.MustWithDigester(object.GetTai())
		defer merkle_ids.PutBlobId(digest)

		if len(object.Metadata.RepoPubkey) == 0 {
			err = errors.ErrorWithStackf(
				"RepoPubkey missing for %s. Fields: %#v",
				sku.String(object),
				object.Metadata.Fields,
			)
			return
		}

		if object.Metadata.GetContentSig().IsNull() {
			err = errors.ErrorWithStackf(
				"signature missing for %s. Fields: %#v",
				sku.String(object),
				object.Metadata.Fields,
			)
			return
		}

		if err = merkle.VerifySignature(
			object.Metadata.RepoPubkey,
			digest.GetBytes(),
			object.Metadata.GetContentSig().GetBytes(),
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
