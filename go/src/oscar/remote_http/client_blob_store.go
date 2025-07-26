package remote_http

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

func (client *client) GetBlobStore() interfaces.BlobStore {
	return client
}

func (client *client) HasBlob(sh interfaces.BlobId) (ok bool) {
	var request *http.Request

	{
		var err error

		if request, err = http.NewRequestWithContext(
			client.GetEnv(),
			"HEAD",
			"/blobs",
			strings.NewReader(digests.Format(sh.GetBlobId())),
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

func (client *client) BlobReader(
	sh interfaces.BlobId,
) (reader interfaces.ReadCloseBlobIdGetter, err error) {
	var request *http.Request

	if request, err = http.NewRequestWithContext(
		client.GetEnv(),
		"GET",
		fmt.Sprintf("/blobs/%s", digests.Format(sh.GetBlobId())),
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
			BlobIdGetter: sh,
		}

	case response.StatusCode >= 300:
		err = ReadErrorFromBody(response)

	default:
		reader = digests.MakeReadCloser(sha.Env{}, response.Body)
	}

	return
}

func (client *client) WriteBlobToRemote(
	localBlobStore interfaces.BlobStore,
	expected *sha.Sha,
) (err error) {
	var actual sha.Sha

	// Closed by the http client's transport (our roundtripper calling
	// request.Write)
	var reader interfaces.ReadCloseBlobIdGetter

	if reader, err = localBlobStore.BlobReader(
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

	if request, err = http.NewRequestWithContext(
		client.GetEnv(),
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

	var shString strings.Builder

	if _, err = io.Copy(&shString, response.Body); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = response.Body.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = actual.Set(strings.TrimSpace(shString.String())); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = expected.AssertEqualsShaLike(&actual); err != nil {
		err = errors.Wrap(err)
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

func (client *client) AllBlobs() interfaces.SeqError[interfaces.BlobId] {
	panic(errors.Err501NotImplemented)
}

func (client *client) Mover() (interfaces.Mover, error) {
	panic(errors.Err501NotImplemented)
}

func (client *client) BlobWriter() (interfaces.WriteCloseBlobIdGetter, error) {
	panic(errors.Err501NotImplemented)
}
