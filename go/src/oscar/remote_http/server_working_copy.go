package remote_http

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

func (server *Server) writeInventoryListLocalWorkingCopy(
	repo *local_working_copy.Repo,
	request Request,
	listSku *sku.Transacted,
) (response Response) {
	listSkuType := builtin_types.GetOrPanic(
		server.Repo.GetImmutableConfigPublic().Blob.GetInventoryListTypeString(),
	).Type

	blobStore := server.Repo.GetBlobStore()

	if listSku != nil {
		if listSku.GetGenre() != genres.InventoryList {
			response.Error(genres.MakeErrUnsupportedGenre(listSku.GetGenre()))
			return
		}

		if blobStore.HasBlob(listSku.GetBlobSha()) {
			response.StatusCode = http.StatusFound
			return
		}

		listSkuType = listSku.GetType()
	}

	typedInventoryListStore := server.Repo.GetTypedInventoryListBlobStore()

	var blobWriter sha.WriteCloser

	{
		var err error

		if blobWriter, err = blobStore.BlobWriter(); err != nil {
			response.Error(err)
			return
		}
	}

	var list *sku.List

	{
		var err error

		if list, err = typedInventoryListStore.ReadInventoryListBlob(
			listSkuType,
			bufio.NewReader(io.TeeReader(request.Body, blobWriter)),
		); err != nil {
			response.Error(err)
			return
		}
	}

	responseBuffer := bytes.NewBuffer(nil)

	// TODO make option to read from headers
	// TODO add remote blob store
	importerOptions := sku.ImporterOptions{
		// TODO
		CheckedOutPrinter: repo.PrinterCheckedOutConflictsForRemoteTransfers(),
	}

	if request.Headers.Get("x-dodder-remote_transfer_options-allow_merge_conflicts") == "true" {
		importerOptions.AllowMergeConflicts = true
	}

	listFormat := server.Repo.GetInventoryListStore().FormatForVersion(
		repo.GetImmutableConfigPublic().Blob.GetStoreVersion(),
	)

	listMissingSkus := sku.MakeList()
	var requestRetry bool

	importerOptions.BlobCopierDelegate = func(
		result sku.BlobCopyResult,
	) (err error) {
		server.Repo.GetEnv().ContinueOrPanicOnDone()

		if result.N != -1 {
			return
		}

		if result.Transacted.GetGenre() == genres.InventoryList {
			requestRetry = true
		}

		listMissingSkus.Add(result.Transacted.CloneTransacted())

		return
	}

	importer := server.Repo.MakeImporter(
		importerOptions,
		sku.GetStoreOptionsRemoteTransfer(),
	)

	if err := server.Repo.ImportList(
		list,
		importer,
	); err != nil {
		if env_dir.IsErrBlobMissing(err) {
			requestRetry = true
		} else {
			response.Error(err)
			return
		}
	}

	bufferedWriter := ohio.BufferedWriter(responseBuffer)
	defer pool.GetBufioWriter().Put(bufferedWriter)

	if _, err := listFormat.WriteInventoryListBlob(
		listMissingSkus,
		bufferedWriter,
	); err != nil {
		response.Error(err)
		return
	}

	if err := bufferedWriter.Flush(); err != nil {
		response.Error(err)
		return
	}

	if requestRetry {
		response.StatusCode = http.StatusExpectationFailed
	} else {
		response.StatusCode = http.StatusCreated
	}

	response.Body = io.NopCloser(responseBuffer)

	return
}
