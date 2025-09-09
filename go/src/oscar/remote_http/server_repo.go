package remote_http

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/tridex"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/india/log_remote_inventory_lists"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (server *Server) writeInventoryList(
	request Request,
	listObject *sku.Transacted,
) (response Response) {
	logRemoteInventoryLists := log_remote_inventory_lists.Make(
		request.ctx,
		server.Repo.GetEnvRepo(),
	)

	if listObject == nil {
		panic("nil list object")
	}

	if listObject.GetGenre() != genres.InventoryList {
		response.Error(genres.MakeErrUnsupportedGenre(listObject.GetGenre()))
		return
	}

	blobStore := server.Repo.GetBlobStore()

	if blobStore.HasBlob(listObject.GetBlobDigest()) {
		response.StatusCode = http.StatusFound
		return
	}

	expected := listObject.GetBlobDigest()

	pubBase64 := request.request.Header.Get(headerRepoPublicKey)

	var logEntry log_remote_inventory_lists.Entry

	if pubBase64 != "" {
		{
			var err error

			var bites []byte

			if bites, err = base64.URLEncoding.DecodeString(pubBase64); err != nil {
				response.Error(err)
				return
			}

			if err = logEntry.SetMerkleId(markl.TypeIdEd25519Pub, bites); err != nil {
				response.Error(err)
				return
			}
		}

		logEntry.EntryType = log_remote_inventory_lists.EntryTypeReceived
		logEntry.Transacted = listObject

		var sig markl.Id

		if err := sig.Set(request.request.Header.Get(headerRepoSig)); err != nil {
			response.Error(err)
			return
		}

		if err := markl.Verify(
			logEntry.PublicKey,
			expected,
			sig,
		); err != nil {
			response.Error(err)
			return
		}
	}

	if len(logEntry.PublicKey.GetBytes()) > 0 {
		if err := logRemoteInventoryLists.Exists(
			logEntry,
		); collections.IsErrNotFound(err) && err != nil {
			err = nil
		} else if err != nil {
			response.Error(err)
			return
		} else {
			response.StatusCode = http.StatusFound
			return
		}
	}

	typedInventoryListStore := server.Repo.GetInventoryListCoderCloset()

	var blobWriter interfaces.BlobWriter

	{
		var err error

		if blobWriter, err = blobStore.MakeBlobWriter(nil); err != nil {
			response.Error(err)
			return
		}
	}

	seqInventoryListSkus := typedInventoryListStore.IterInventoryListBlobSkusFromReader(
		listObject.GetType(),
		bufio.NewReader(io.TeeReader(request.Body, blobWriter)),
	)

	b := bytes.NewBuffer(nil)
	writtenNeededBlobs := tridex.Make()

	{
		count := 0

		for sk, err := range seqInventoryListSkus {
			errors.ContextContinueOrPanic(server.Repo.GetEnv())

			if err != nil {
				response.Error(err)
				return
			}

			blobSha := sk.GetBlobDigest()

			var ok bool
			ok, err = server.blobCache.HasBlob(blobSha)
			if err != nil {
				response.Error(err)
				return
			}

			blobShaString := markl.Format(blobSha)

			if ok || writtenNeededBlobs.ContainsExpansion(blobShaString) {
				continue
			}

			ui.Log().Printf("missing blob: %s", blobSha)

			fmt.Fprintf(b, "%s\n", blobSha)
			writtenNeededBlobs.Add(blobShaString)
			count++
		}

		ui.Err().Printf("missing blobs: %d", count)
	}

	if err := blobWriter.Close(); err != nil {
		response.Error(err)
		return
	}

	actual := blobWriter.GetMarklId()

	if err := markl.MakeErrNotEqual(expected, actual); err != nil {
		ui.Err().Printf(
			"received list has different sha: expected: %s, actual: %s",
			expected,
			actual,
		)

		// response.ErrorWithStatus(http.StatusBadRequest, err)
		// return
	}

	ui.Log().Printf("list sha matches: %s", expected)

	// TODO make merge conflicts impossible

	response.StatusCode = http.StatusCreated
	response.Body = io.NopCloser(b)

	if err := server.Repo.GetObjectStore().Commit(
		listObject,
		sku.CommitOptions{},
	); err != nil {
		response.Error(err)
		return
	}

	if len(logEntry.PublicKey.GetBytes()) > 0 {
		if err := logRemoteInventoryLists.Append(
			logEntry,
		); err != nil {
			response.Error(err)
			return
		}
	}

	return
}
