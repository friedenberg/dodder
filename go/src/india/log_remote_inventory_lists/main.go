package log_remote_inventory_lists

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type EntryType interface {
	entryType()
}

//go:generate stringer -type=entryType
type entryType byte

func (entryType) entryType() {}

const (
	EntryTypeSent = entryType(iota)
	EntryTypeReceived
)

type Entry struct {
	EntryType
	repo_signing.PublicKey
	*sku.Transacted
}

type Log interface {
	errors.Flusher
	initialize(errors.Context, env_repo.Env)
	Key(Entry) (string, error)
	Append(Entry) error
	Exists(Entry) error
}

func Make(ctx errors.Context, envRepo env_repo.Env) (log Log) {
	sv := envRepo.GetConfigPrivate().Blob.GetStoreVersion()

	if store_version.Less(sv, store_version.V8) {
		ctx.CancelWithErrorf("unsupported store version: %s")
		return nil
	}

	log = &v0{}

	log.initialize(ctx, envRepo)
	ctx.After(log.Flush)

	return
}
