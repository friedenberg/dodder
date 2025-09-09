package remote_http

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

func (server *Server) writeInventoryListTypedBlobLocalWorkingCopy(
	repo *local_working_copy.Repo,
	request Request,
) (response Response) {
	listCoderCloset := server.Repo.GetInventoryListCoderCloset()

	responseBuffer := bytes.NewBuffer(nil)

	// TODO make option to read from headers
	// TODO add remote blob store
	importerOptions := sku.ImporterOptions{
		// TODO
		CheckedOutPrinter: repo.PrinterCheckedOutConflictsForRemoteTransfers(),
	}

	if request.Headers.Get(
		"x-dodder-remote_transfer_options-allow_merge_conflicts",
	) == "true" {
		importerOptions.AllowMergeConflicts = true
	}

	listMissingSkus := sku.MakeListTransacted()
	var requestRetry bool

	importerOptions.BlobCopierDelegate = func(
		result sku.BlobCopyResult,
	) (err error) {
		errors.ContextContinueOrPanic(server.Repo.GetEnv())

		if result.N != -1 {
			return
		}

		if result.Transacted.GetGenre() == genres.InventoryList {
			requestRetry = true
		}

		ui.Log().Print(
			"missing blob for list: %s",
			sku.String(result.Transacted),
		)

		listMissingSkus.Add(result.Transacted.CloneTransacted())

		return
	}

	importer := server.Repo.MakeImporter(
		importerOptions,
		sku.GetStoreOptionsRemoteTransfer(),
	)

	seq := listCoderCloset.AllDecodedObjectsFromStream(
		bufio.NewReader(request.Body),
		nil,
	)

	if err := server.Repo.ImportSeq(
		seq,
		importer,
	); err != nil {
		if env_dir.IsErrBlobMissing(err) {
			requestRetry = true
		} else {
			response.Error(err)
			return
		}
	}

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(
		responseBuffer,
	)
	defer repoolBufferedWriter()

	listType := ids.GetOrPanic(
		repo.GetImmutableConfigPublic().GetInventoryListTypeId(),
	).Type

	inventoryListCoderCloset := server.Repo.GetInventoryListCoderCloset()

	if _, err := inventoryListCoderCloset.WriteBlobToWriter(
		repo,
		listType,
		quiter.MakeSeqErrorFromSeq(listMissingSkus.All()),
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

func (server *Server) writeInventoryListLocalWorkingCopy(
	repo *local_working_copy.Repo,
	request Request,
	listSku *sku.Transacted,
) (response Response) {
	listSkuType := ids.GetOrPanic(
		server.Repo.GetImmutableConfigPublic().GetInventoryListTypeId(),
	).Type

	blobStore := server.Repo.GetBlobStore()

	if listSku != nil {
		if listSku.GetGenre() != genres.InventoryList {
			response.Error(genres.MakeErrUnsupportedGenre(listSku.GetGenre()))
			return
		}

		if blobStore.HasBlob(listSku.GetBlobDigest()) {
			response.StatusCode = http.StatusFound
			return
		}

		listSkuType = listSku.GetType()
	}

	listCoderCloset := server.Repo.GetInventoryListCoderCloset()

	var blobWriter interfaces.WriteCloseMarklIdGetter

	{
		var err error

		if blobWriter, err = blobStore.MakeBlobWriter(""); err != nil {
			response.Error(err)
			return
		}
	}

	var list *sku.ListTransacted

	{
		var err error

		if list, err = listCoderCloset.ReadInventoryListBlob(
			repo,
			listSkuType,
			bufio.NewReader(io.TeeReader(request.Body, blobWriter)),
		); err != nil {
			response.Error(err)
			return
		}
	}

	ui.Log().Printf("read list: %d objects", list.Len())

	responseBuffer := bytes.NewBuffer(nil)

	// TODO make option to read from headers
	// TODO add remote blob store
	importerOptions := sku.ImporterOptions{
		// TODO
		CheckedOutPrinter: repo.PrinterCheckedOutConflictsForRemoteTransfers(),
	}

	if request.Headers.Get(
		"x-dodder-remote_transfer_options-allow_merge_conflicts",
	) == "true" {
		importerOptions.AllowMergeConflicts = true
	}

	listMissingSkus := sku.MakeListTransacted()
	var requestRetry bool

	importerOptions.BlobCopierDelegate = func(
		result sku.BlobCopyResult,
	) (err error) {
		errors.ContextContinueOrPanic(server.Repo.GetEnv())

		if result.N != -1 {
			return
		}

		if result.Transacted.GetGenre() == genres.InventoryList {
			requestRetry = true
		}

		ui.Log().Print(
			"missing blob for list: %s",
			sku.String(result.Transacted),
		)

		listMissingSkus.Add(result.Transacted.CloneTransacted())

		return
	}

	importer := server.Repo.MakeImporter(
		importerOptions,
		sku.GetStoreOptionsRemoteTransfer(),
	)

	if err := server.Repo.ImportSeq(
		quiter.MakeSeqErrorFromSeq(list.All()),
		importer,
	); err != nil {
		if env_dir.IsErrBlobMissing(err) {
			requestRetry = true
		} else {
			response.Error(err)
			return
		}
	}

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(
		responseBuffer,
	)
	defer repoolBufferedWriter()

	inventoryListCoderCloset := repo.GetInventoryListCoderCloset()

	if _, err := inventoryListCoderCloset.WriteTypedBlobToWriter(
		repo,
		ids.GetOrPanic(repo.GetImmutableConfigPublic().GetInventoryListTypeId()).Type,
		quiter.MakeSeqErrorFromSeq(listMissingSkus.All()),
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
