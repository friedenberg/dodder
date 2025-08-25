package log_remote_inventory_lists

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
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
	merkle.PublicKey
	*sku.Transacted
}

type Log interface {
	errors.Flusher
	initialize(interfaces.Context, env_repo.Env)
	Key(Entry) (string, error)
	Append(Entry) error
	Exists(Entry) error
}

func Make(ctx interfaces.Context, envRepo env_repo.Env) (log Log) {
	sv := envRepo.GetConfigPrivate().Blob.GetStoreVersion()

	if store_version.Less(sv, store_version.V8) {
		errors.ContextCancelWithErrorf(ctx, "unsupported store version: %s", sv)
		return nil
	}

	log = &v0{}

	log.initialize(ctx, envRepo)
	ctx.After(errors.MakeFuncContextFromFuncErr(log.Flush))

	return
}
