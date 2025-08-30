package log_remote_inventory_lists

import (
	"bufio"
	"encoding/base64"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/tridex"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

// TODO consider moving this or refactoring it as it's currently not really used
type v0 struct {
	once      sync.Once
	path      string
	lockSmith interfaces.LockSmith
	file      *os.File
	values    interfaces.MutableTridex
}

func (log *v0) Flush() (err error) {
	if _, err = log.file.Seek(0, io.SeekStart); err != nil {
		err = errors.Wrap(err)
		return
	}

	bufferedWriter := bufio.NewWriter(log.file)

	enc := gob.NewEncoder(bufferedWriter)

	if err = enc.Encode(log.values); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = log.file.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = log.lockSmith.Unlock(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return nil
}

func (log *v0) initialize(ctx interfaces.Context, env env_repo.Env) {
	gob.Register(tridex.Make())

	log.values = tridex.Make()

	dir := env.DirCacheInventoryListLog()

	log.path = filepath.Join(dir, "log-v0")
	log.lockSmith = file_lock.New(
		env,
		filepath.Join(dir, "log-v0.lock"),
		"log_remote_inventory_lists",
	)

	if err := log.lockSmith.Lock(); err != nil {
		ctx.Cancel(err)
		return
	}

	{
		var err error

		if log.file, err = files.TryOrMakeDirIfNecessary(
			log.path,
			files.OpenCreate,
		); err != nil {
			ctx.Cancel(err)
			return
		}
	}
}

func (log *v0) Append(entry Entry) (err error) {
	if err = log.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var key string

	if key, err = log.Key(entry); err != nil {
		err = errors.Wrap(err)
		return
	}

	log.values.Add(key)

	return
}

func (log *v0) Key(entry Entry) (key string, err error) {
	if entry.EntryType == nil {
		err = errors.ErrorWithStackf("nil entry type")
		return
	}

	digest, repool := markl.HashTypeSha256.FromStringFormat(
		"%s%s%s%s",
		entry.EntryType,
		base64.URLEncoding.EncodeToString(entry.PublicKey),
		entry.GetObjectId(),
		entry.GetBlobDigest(),
	)

	defer repool()

	key = markl.Format(digest)

	return
}

func (log *v0) Exists(entry Entry) (err error) {
	if err = log.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var key string

	if key, err = log.Key(entry); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !log.values.ContainsExpansion(key) {
		return collections.ErrNotFound
	}

	return
}

func (log *v0) readIfNecessary() (err error) {
	log.once.Do(
		func() {
			bufferedReader := bufio.NewReader(log.file)

			dec := gob.NewDecoder(bufferedReader)

			if err = dec.Decode(log.values); err != nil {
				if errors.IsEOF(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
				}

				return
			}
		},
	)

	return
}
