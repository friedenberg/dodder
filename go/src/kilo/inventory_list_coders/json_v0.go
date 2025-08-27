package inventory_list_coders

import (
	"bufio"
	"encoding/json"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_json_fmt"
)

type jsonV0 struct {
	genesisConfig genesis_configs.ConfigPrivate
}

func (coder jsonV0) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if err = object.Verify(); err != nil {
		err = errors.Wrapf(
			err,
			"Sku: %s, Tags %s",
			sku.String(object),
			quiter.StringCommaSeparated(object.Metadata.GetTags()),
		)
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

	bites, err := bufferedReader.ReadBytes('\n')
	if err != nil {
		return
	}

	if err = json.Unmarshal(bites, &objectJson); err != nil {
		err = errors.Wrapf(err, "Line: %q", bites)
		return
	}

	if err = objectJson.ToTransacted(object, nil); err != nil {
		err = errors.Wrapf(err, "Line: %q", bites)
		return
	}

	if object.GetType().String() == ids.TypeInventoryListJsonV0 {
		if err = object.Verify(); err != nil {
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

	return
}
