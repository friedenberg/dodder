package remote_http

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/bravo/comments"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/india/log_remote_inventory_lists"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (client client) WriteInventoryListObject(t *sku.Transacted) (err error) {
	return comments.Implement()
}

// TODO add progress bar
func (client client) ImportInventoryList(
	blobStore interfaces.BlobStore,
	listSku *sku.Transacted,
) (err error) {
	logEntry := log_remote_inventory_lists.Entry{
		EntryType:  log_remote_inventory_lists.EntryTypeSent,
		PublicKey:  client.configImmutable.Blob.GetPublicKey(),
		Transacted: listSku,
	}

	if err = client.logRemoteInventoryLists.Exists(
		logEntry,
	); collections.IsErrNotFound(err) {
		err = nil
	} else if err != nil {
		err = errors.Wrap(err)
		return
	} else {
		return
	}

	ui.Log().Printf("importing list: %s", sku.String(listSku))

	var sbListSkuBox strings.Builder

	{
		bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(
			&sbListSkuBox,
		)
		defer repoolBufferedWriter()

		if _, err = client.inventoryListCoderCloset.WriteObjectToWriter(
			listSku.GetType(),
			listSku,
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = bufferedWriter.Flush(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var listBlobReader io.ReadCloser

	if listBlobReader, err = blobStore.BlobReader(listSku.GetBlobDigest()); err != nil {
		err = errors.Wrap(err)
		return
	}

	var request *http.Request

	if request, err = client.newRequest(
		"POST",
		fmt.Sprintf(
			"/inventory_lists/%s/%s",
			listSku.GetType(),
			strings.TrimSpace(sbListSkuBox.String()),
		),
		listBlobReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	{
		pubKey := client.repo.GetImmutableConfigPublic().GetPublicKey()

		request.Header.Add(
			headerRepoPublicKey,
			base64.URLEncoding.EncodeToString(pubKey),
		)
	}

	{
		sig := blech32.Value{
			HRP: merkle.HRPObjectSigV1,
		}

		key := client.repo.GetImmutableConfigPrivate().Blob.GetPrivateKey()

		if sig.Data, err = merkle.SignBytes(
			key,
			listSku.GetBlobDigest().GetBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		request.Header.Add(headerRepoSig, sig.String())
	}

	// TODO ensure that conflicts were addressed prior to importing
	// if options.AllowMergeConflicts {
	// 	request.Header.Add("x-dodder-remote_transfer_options-allow_merge_conflicts",
	// "true")
	// }

	var response *http.Response

	if response, err = client.http.Do(request); err != nil {
		err = errors.ErrorWithStackf("failed to read response: %w", err)
		return
	}

	if err = ReadErrorFromBodyOnNot(
		response,
		http.StatusCreated,
		http.StatusFound,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var digests merkle.Slice

	if _, err = digests.ReadFrom(bufio.NewReader(response.Body)); err != nil {
		err = errors.Wrap(err)
		return
	}

	if len(digests) > 0 {
		ui.Err().Printf("sending blobs: %d", len(digests))
	}

	for _, digest := range digests {
		if err = client.WriteBlobToRemote(blobStore, digest); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = response.Body.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = client.logRemoteInventoryLists.Append(
		logEntry,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (client client) ReadLast() (max *sku.Transacted, err error) {
	return nil, comments.Implement()
}

func (client client) IterInventoryList(
	blobSha interfaces.BlobId,
) interfaces.SeqError[*sku.Transacted] {
	return nil
}

func (client client) ReadAllSkus(
	f func(besty, sk *sku.Transacted) error,
) (err error) {
	return comments.Implement()
}

func (client client) AllInventoryListObjects() interfaces.SeqError[*sku.Transacted] {
	var request *http.Request

	{
		var err error

		if request, err = client.newRequest(
			"GET",
			"/inventory_lists",
			nil,
		); err != nil {
			client.envUI.Cancel(err)
			return nil
		}
	}

	var response *http.Response

	{
		var err error

		if response, err = client.http.Do(request); err != nil {
			errors.ContextCancelWithErrorAndFormat(
				client.envUI,
				err,
				"failed to read response",
			)
			return nil
		}
	}

	return client.inventoryListCoderCloset.AllDecodedObjectsFromStream(
		response.Body,
	)
}
