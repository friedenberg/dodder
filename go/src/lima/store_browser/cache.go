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

func (c *Store) getCachePath() string {
	return path.Join(c.externalStoreInfo.DirCache, "tab_cache")
}

func (c *Store) initializeCache() (err error) {
	c.tabCache.Rows = make(map[string]browser_items.ItemId)

	var f *os.File

	if f, err = files.OpenExclusiveReadOnly(
		c.getCachePath(),
	); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	defer errors.DeferredCloser(&err, f)

	br := bufio.NewReader(f)
	dec := gob.NewDecoder(br)

	if err = dec.Decode(&c.tabCache); err != nil {
		ui.Err().Printf("browser tab cache parse failed: %s", err)
		err = nil
		return
	}

	return
}

func (c *Store) resetCacheIfNecessary(
	resp *http.Response,
) (err error) {
	if resp == nil {
		return
	}

	timeRaw := resp.Header.Get("X-Chrest-Startup-Time")

	var newLaunchTime ids.Tai

	if err = newLaunchTime.SetFromRFC3339(timeRaw); err != nil {
		err = errors.Wrap(err)
		return
	}

	if newLaunchTime.Equals(c.tabCache.LaunchTime) {
		return
	}

	c.tabCache.LaunchTime = newLaunchTime
	clear(c.tabCache.Rows)

	return
}

func (c *Store) flushCache() (err error) {
	var file *os.File

	if file, err = files.OpenExclusiveWriteOnly(
		c.getCachePath(),
	); err != nil {
		if errors.IsNotExist(err) {
			if file, err = files.TryOrMakeDirIfNecessary(
				c.getCachePath(),
				files.CreateExclusiveWriteOnly,
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	defer errors.DeferredCloser(&err, file)

	bw := bufio.NewWriter(file)
	defer errors.DeferredFlusher(&err, bw)

	dec := gob.NewEncoder(bw)

	if err = dec.Encode(&c.tabCache); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
