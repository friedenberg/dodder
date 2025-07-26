package remote_http

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/tridex"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
)

type serverBlobCache struct {
	ui             fd.Std
	localBlobStore interfaces.BlobStore
	shas           interfaces.MutableTridex
	init           sync.Once
}

func (serverBlobCache *serverBlobCache) populate() (err error) {
	serverBlobCache.shas = tridex.Make()

	{
		count := 0

		for sh, errIter := range serverBlobCache.localBlobStore.AllBlobs() {
			if errIter != nil {
				err = errors.Wrap(errIter)
				return
			}

			serverBlobCache.shas.Add(digests.Format(sh))
			count++
		}

		ui.Log().Printf("have blobs: %d", count)
	}

	return
}

func (serverBlobCache *serverBlobCache) HasBlob(
	blobSha interfaces.BlobId,
) (ok bool, err error) {
	serverBlobCache.init.Do(
		func() {
			if err = serverBlobCache.populate(); err != nil {
				err = errors.Wrap(err)
			}
		},
	)

	if err != nil {
		return
	}

	if serverBlobCache.shas.ContainsExpansion(
		digests.Format(blobSha),
	) {
		ok = true
		return
	}

	return
}
