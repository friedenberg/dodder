package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/todo"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type directoryPaths interface {
	interfaces.DirectoryPaths
	init(sv config_immutable.StoreVersion, xdg xdg.XDG) error
}

type directoryV0 struct {
	sv       config_immutable.StoreVersion
	basePath string
}

func (c *directoryV0) init(
	sv config_immutable.StoreVersion,
	xdg xdg.XDG,
) (err error) {
	c.sv = sv
	return
}

func (c directoryV0) GetDirectoryPaths() interfaces.DirectoryPaths {
	return c
}

func (c directoryV0) FileCacheDormant() string {
	return c.DirZit("Schlummernd")
}

func (c directoryV0) FileTags() string {
	return c.DirZit("Etiketten")
}

func (c directoryV0) FileLock() string {
	return c.DirZit("Lock")
}

func (c directoryV0) FileConfigPermanent() string {
	return c.DirZit("KonfigAngeboren")
}

func (c directoryV0) FileConfigMutable() string {
	return c.DirZit("KonfigErworben")
}

func (s directoryV0) Dir(p ...string) string {
	return filepath.Join(stringSliceJoin(s.basePath, p)...)
}

func (s directoryV0) DirZit(p ...string) string {
	return s.Dir(stringSliceJoin(".zit", p)...)
}

func (s directoryV0) DirCache(p ...string) string {
	return s.DirZit(append([]string{"Verzeichnisse"}, p...)...)
}

func (s directoryV0) DirCacheRepo(p ...string) string {
	return s.DirZit(append([]string{"Verzeichnisse", "Kasten"}, p...)...)
}

func (s directoryV0) DirCacheDurable(p ...string) string {
	return s.DirZit(append([]string{"VerzeichnisseDurable"}, p...)...)
}

func (s directoryV0) DirObjects(p ...string) string {
	return s.DirZit(append([]string{"Objekten2"}, p...)...)
}

func (s directoryV0) DirLostAndFound() string {
	return s.DirZit("Verloren+Gefunden")
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
	return s.DirZit("Kennung")
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
