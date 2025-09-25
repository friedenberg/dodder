package store_browser

import (
	"bufio"
	"encoding/gob"
	"net/http"
	"os"
	"path"

	"code.linenisgreat.com/chrest/go/src/charlie/browser_items"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type cache struct {
	LaunchTime ids.Tai
	Rows       map[string]browser_items.ItemId // map[browserItem.ExternalId]browserItemId
}

func (store *Store) getCachePath() string {
	return path.Join(store.externalStoreInfo.DirCache, "tab_cache")
}

func (store *Store) initializeCache() (err error) {
	store.tabCache.Rows = make(map[string]browser_items.ItemId)

	var f *os.File

	if f, err = files.OpenExclusiveReadOnly(
		store.getCachePath(),
	); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	defer errors.DeferredCloser(&err, f)

	br := bufio.NewReader(f)
	dec := gob.NewDecoder(br)

	if err = dec.Decode(&store.tabCache); err != nil {
		ui.Err().Printf("browser tab cache parse failed: %s", err)
		err = nil
		return err
	}

	return err
}

func (store *Store) resetCacheIfNecessary(
	resp *http.Response,
) (err error) {
	if resp == nil {
		return err
	}

	timeRaw := resp.Header.Get("X-Chrest-Startup-Time")

	var newLaunchTime ids.Tai

	if err = newLaunchTime.SetFromRFC3339(timeRaw); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if newLaunchTime.Equals(store.tabCache.LaunchTime) {
		return err
	}

	store.tabCache.LaunchTime = newLaunchTime
	clear(store.tabCache.Rows)

	return err
}

func (store *Store) flushCache() (err error) {
	var file *os.File

	if file, err = files.OpenExclusiveWriteOnly(
		store.getCachePath(),
	); err != nil {
		if errors.IsNotExist(err) {
			if file, err = files.TryOrMakeDirIfNecessary(
				store.getCachePath(),
				files.CreateExclusiveWriteOnly,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	defer errors.DeferredCloser(&err, file)

	bw := bufio.NewWriter(file)
	defer errors.DeferredFlusher(&err, bw)

	dec := gob.NewEncoder(bw)

	if err = dec.Encode(&store.tabCache); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
