package store_browser

import (
	"context"
	"net/url"
	"syscall"

	"code.linenisgreat.com/chrest/go/src/charlie/browser_items"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

func (s *Store) Initialize(esi store_workspace.Supplies) (err error) {
	s.externalStoreInfo = esi

	if err = s.browser.Read(); err != nil {
		err = errors.Wrap(err)
		return
	}

	wg := errors.MakeWaitGroupParallel()

	wg.Do(s.initializeUrls)
	wg.Do(s.initializeIndex)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Store) initializeUrls() (err error) {
	var req browser_items.BrowserRequestGet
	var resp browser_items.HTTPResponseWithRequestPayloadGet

	ui.Log().Print("getting all")

	ctx := context.Background()
	ctxWithTimeout, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	if resp, err = s.browser.GetAll(
		ctxWithTimeout,
		req,
	); err != nil {
		if errors.IsErrno(err, syscall.ECONNREFUSED) {
			if !s.config.GetConfig().Quiet {
				ui.Err().Print("chrest offline")
			}

			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	ui.Log().Print("got all")

	s.urls = make(map[url.URL][]Item, len(resp.RequestPayloadGet))

	if err = s.resetCacheIfNecessary(resp.Response); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, item := range resp.RequestPayloadGet {
		i := Item{Item: item}

		u := i.Url.Url()

		s.urls[u] = append(s.urls[u], i)
		s.itemsById[i.GetObjectId().String()] = i
	}

	return
}
