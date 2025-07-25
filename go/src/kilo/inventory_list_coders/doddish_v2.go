package inventory_list_coders

import (
	"bufio"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type DoddishV2 struct {
	Box                    *box_format.BoxTransacted
	ImmutableConfigPrivate genesis_configs.ConfigPrivate
}

func (coder DoddishV2) GetType() ids.Type {
	return ids.MustType(ids.TypeInventoryListV2)
}

func (coder DoddishV2) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if object.Metadata.GetDigest().IsNull() {
		err = errors.ErrorWithStackf("empty sha: %q", sku.String(object))
		return
	}

	if object.Metadata.RepoSig.IsEmpty() {
		err = errors.ErrorWithStackf("no repo signature")
		return
	}

	var n1 int64
	var n2 int

	n1, err = coder.Box.EncodeStringTo(object, bufferedWriter)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n2, err = fmt.Fprintf(bufferedWriter, "\n")
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coder DoddishV2) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var isEOF bool

	if n, err = coder.Box.ReadStringFormat(object, bufferedReader); err != nil {
		if err == io.EOF {
			isEOF = true

			if n == 0 {
				return
			}
		} else {
			err = errors.Wrap(err)
			return
		}
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

	if isEOF {
		err = io.EOF
	}

	return
}
