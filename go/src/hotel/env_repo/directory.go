package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/todo"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type directoryPaths interface {
	interfaces.DirectoryPaths
	init(sv genesis_config.StoreVersion, xdg xdg.XDG) error
}

type directoryV0 struct {
	sv       genesis_config.StoreVersion
	basePath string
}

func (c *directoryV0) init(
	sv genesis_config.StoreVersion,
	xdg xdg.XDG,
) (err error) {
	c.sv = sv
	return
}

func (c directoryV0) GetDirectoryPaths() interfaces.DirectoryPaths {
	return c
}

func (c directoryV0) FileCacheDormant() string {
	return c.DirDodder("Schlummernd")
}

func (c directoryV0) FileTags() string {
	return c.DirDodder("Etiketten")
}

func (c directoryV0) FileLock() string {
	return c.DirDodder("Lock")
}

func (c directoryV0) FileConfigPermanent() string {
	return c.DirDodder("KonfigAngeboren")
}

func (c directoryV0) FileConfigMutable() string {
	return c.DirDodder("KonfigErworben")
}

func (s directoryV0) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(s.basePath, p)...)
}

func (s directoryV0) DirDodder(p ...string) string {
	return s.Dir(stringSliceJoin(".zit", p)...)
}

func (s directoryV0) DirCache(p ...string) string {
	return s.DirDodder(append([]string{"Verzeichnisse"}, p...)...)
}

func (s directoryV0) DirCacheRepo(p ...string) string {
	return s.DirDodder(append([]string{"Verzeichnisse", "Kasten"}, p...)...)
}

func (s directoryV0) DirCacheDurable(p ...string) string {
	return s.DirDodder(append([]string{"VerzeichnisseDurable"}, p...)...)
}

func (s directoryV0) DirObjects(p ...string) string {
	return s.DirDodder(append([]string{"Objekten2"}, p...)...)
}

func (s directoryV0) DirLostAndFound() string {
	return s.DirDodder("Verloren+Gefunden")
}

func (s directoryV0) DirCacheObjects() string {
	return s.DirCache("Objekten")
}

func (s directoryV0) DirCacheObjectPointers() string {
	return s.DirCache("Verweise")
}

func (s directoryV0) DirCacheInventoryListLog() string {
	return s.DirCache("inventory_list_logs")
}

func (s directoryV0) DirObjectId() string {
	return s.DirDodder("Kennung")
}

func (s directoryV0) FileCacheObjectId() string {
	return s.DirCache("Kennung")
}

func (s directoryV0) DirInventoryLists() string {
	return s.DirObjects("inventory_lists")
}

func (s directoryV0) DirBlobs() string {
	return s.DirObjects("blobs")
}

func (s directoryV0) FileInventoryListLog() string {
	panic(todo.Implement())
}
