package remote_http

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"iter"
	"net/http"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/todo"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/india/log_remote_inventory_lists"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (client client) FormatForVersion(
	sv interfaces.StoreVersion,
) sku.ListFormat {
	return client.localRepo.GetInventoryListStore().FormatForVersion(sv)
}

func (client client) WriteInventoryListObject(t *sku.Transacted) (err error) {
	return todo.Implement()
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
	listFormat := client.GetInventoryListStore().FormatForVersion(
		client.localRepo.GetImmutableConfigPublic().GetStoreVersion(),
	)

	buffer := bytes.NewBuffer(nil)

	var list *sku.List

	// TODO add support for "broken" inventory lists that have unstable sorts
	if list, err = sku.CollectList(
		client.typedBlobStore.IterInventoryListBlobSkusFromBlobStore(
			listSku.GetType(),
			blobStore,
			listSku.GetBlobSha(),
		),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	ui.Log().Printf("collected list (%d): %s", list.Len(), sku.String(listSku))

	{
		bufferedWriter := ohio.BufferedWriter(buffer)
		defer pool.GetBufioWriter().Put(bufferedWriter)

		// TODO make a reader version of inventory lists to avoid allocation
		if _, err = listFormat.WriteInventoryListBlob(
			list,
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

	var sbListSkuBox strings.Builder

	{
		bufferedWriter := ohio.BufferedWriter(&sbListSkuBox)
		defer pool.GetBufioWriter().Put(bufferedWriter)

		if _, err = client.typedBlobStore.WriteObjectToWriter(
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

	var request *http.Request

	if request, err = http.NewRequestWithContext(
		client.GetEnv(),
		"POST",
		fmt.Sprintf("/inventory_lists/%s", strings.TrimSpace(sbListSkuBox.String())),
		buffer,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	{
		key := client.localRepo.GetImmutableConfigPrivate().Blob.GetPublicKey()

		request.Header.Add(
			headerRepoPublicKey,
			base64.URLEncoding.EncodeToString(key),
		)
	}

	{
		key := client.localRepo.GetImmutableConfigPrivate().Blob.GetPrivateKey()

		var sig string

		if sig, err = repo_signing.SignBase64(
			key,
			listSku.GetBlobSha().GetShaBytes(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		request.Header.Add(headerSha256Sig, sig)
	}

	// TODO ensure that conflicts were addressed prior to importing
	// if options.AllowMergeConflicts {
	// 	request.Header.Add("x-dodder-remote_transfer_options-allow_merge_conflicts", "true")
	// }

	var response *http.Response

	if response, err = client.http.Do(request); err != nil {
		err = errors.ErrorWithStackf("failed to read response: %w", err)
		return
	}

	ui.Log().Printf("sent list (%d): %s", list.Len(), sku.String(listSku))

	if err = ReadErrorFromBodyOnNot(
		response,
		http.StatusCreated,
		http.StatusFound,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var shas sha.Slice

	if _, err = shas.ReadFrom(bufio.NewReader(response.Body)); err != nil {
		err = errors.Wrap(err)
		return
	}

	if len(shas) > 0 {
		ui.Err().Printf("sending blobs: %d", len(shas))
	}

	for _, sh := range shas {
		if err = client.WriteBlobToRemote(blobStore, sh); err != nil {
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
	return nil, todo.Implement()
}

func (client client) IterInventoryList(
	blobSha interfaces.Sha,
) iter.Seq2[*sku.Transacted, error] {
	return nil
}

func (client client) ReadAllSkus(
	f func(besty, sk *sku.Transacted) error,
) (err error) {
	return todo.Implement()
}

func (client client) IterAllInventoryLists() iter.Seq2[*sku.Transacted, error] {
	var request *http.Request

	{
		var err error

		if request, err = http.NewRequestWithContext(
			client.GetEnv(),
			"GET",
			"/inventory_lists",
			nil,
		); err != nil {
			client.envUI.CancelWithError(err)
			return nil
		}
	}

	var response *http.Response

	{
		var err error

		if response, err = client.http.Do(request); err != nil {
			client.envUI.CancelWithErrorAndFormat(err, "failed to read response")
			return nil
		}
	}

	return client.typedBlobStore.AllDecodedObjectsFromStream(response.Body)
}
