package remote_http

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/markl_io"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
)

func (client *client) GetBlobStore() blob_stores.BlobStoreInitialized {
	return blob_stores.BlobStoreInitialized{
		BlobStoreConfigNamed: blob_stores.BlobStoreConfigNamed{
			Name: "remote",
			// TODO populate these
			// BasePath:
			// Config:
		},
		BlobStore: client,
	}
}

func (client *client) GetBlobStoreConfig() blob_store_configs.Config {
	panic(errors.Err501NotImplemented)
}

func (client *client) GetDefaultHashType() interfaces.HashType {
	panic(errors.Err501NotImplemented)
}

func (client *client) HasBlob(blobId interfaces.MarklId) (ok bool) {
	var request *http.Request

	{
		var err error

		if request, err = client.newRequest(
			"HEAD",
			fmt.Sprintf("/blobs/%s", blobId),
			nil,
		); err != nil {
			client.GetEnv().Cancel(err)
		}
	}

	var response *http.Response

	{
		var err error

		if response, err = client.http.Do(request); err != nil {
			client.GetEnv().Cancel(err)
		}
	}

	ok = response.StatusCode == http.StatusNoContent

	return
}

func (client *client) MakeBlobReader(
	blobId interfaces.MarklId,
) (reader interfaces.BlobReader, err error) {
	var request *http.Request

	if request, err = client.newRequest(
		"GET",
		fmt.Sprintf("/blobs/%s", blobId),
		nil,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var response *http.Response

	if response, err = client.http.Do(request); err != nil {
		err = errors.Wrap(err)
		return
	}

	switch {
	case response.StatusCode == http.StatusNotFound:
		err = env_dir.ErrBlobMissing{
			BlobId: blobId,
		}

	case response.StatusCode >= 300:
		err = ReadErrorFromBody(response)

	default:
		var hashType markl.HashType

		if hashType, err = markl.GetHashTypeOrError(
			blobId.GetMarklType().GetMarklTypeId(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		reader = markl_io.MakeReadCloser(
			hashType.Get(),
			response.Body,
		)
	}

	return
}

func (client *client) WriteBlobToRemote(
	localBlobStore interfaces.BlobStore,
	expected interfaces.MarklId,
) (err error) {
	// Closed by the http client's transport (our roundtripper calling
	// request.Write)
	var reader interfaces.BlobReader

	if reader, err = localBlobStore.MakeBlobReader(
		expected,
	); err != nil {
		if env_dir.IsErrBlobMissing(err) {
			// TODO make an option to collect this error at the present it, and
			// an
			// option to fetch it from another remote store
			ui.Err().Printf("Blob missing locally: %q", expected)
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	var request *http.Request

	if request, err = client.newRequest(
		"POST",
		"/blobs",
		reader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	request.TransferEncoding = []string{"chunked"}

	var response *http.Response

	if response, err = client.http.Do(request); err != nil {
		err = errors.ErrorWithStackf("failed to read response: %w", err)
		return
	}

	if err = ReadErrorFromBodyOnNot(response, http.StatusCreated); err != nil {
		err = errors.Wrap(err)
		return
	}

	var digestString strings.Builder

	if _, err = io.Copy(&digestString, response.Body); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = response.Body.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var actual markl.Id

	if err = actual.Set(
		strings.TrimSpace(digestString.String()),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = markl.AssertEqual(expected, &actual); err != nil {
		ui.Debug().Print(err)
		err = errors.Wrapf(err, "Raw Blob Id: %q", digestString.String())
		return
	}

	return
}

//   _   _       _
//  | \ | | ___ | |_
//  |  \| |/ _ \| __|
//  | |\  | (_) | |_
//  |_| \_|\___/ \__|
//
//   ___                 _                           _           _
//  |_ _|_ __ ___  _ __ | | ___ _ __ ___   ___ _ __ | |_ ___  __| |
//   | || '_ ` _ \| '_ \| |/ _ \ '_ ` _ \ / _ \ '_ \| __/ _ \/ _` |
//   | || | | | | | |_) | |  __/ | | | | |  __/ | | | ||  __/ (_| |
//  |___|_| |_| |_| .__/|_|\___|_| |_| |_|\___|_| |_|\__\___|\__,_|
//                |_|

func (client *client) GetBlobStoreDescription() string {
	panic(errors.Err501NotImplemented)
}

func (client *client) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	panic(errors.Err501NotImplemented)
}

func (client *client) AllBlobs() interfaces.SeqError[interfaces.MarklId] {
	panic(errors.Err501NotImplemented)
}

func (client *client) MakeBlobWriter(
	marklHashType interfaces.HashType,
) (interfaces.BlobWriter, error) {
	panic(errors.Err501NotImplemented)
}
