package directory_layout

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type DirectoryLayout interface {
	interfaces.RepoDirectoryLayout
	Initialize(xdg.XDG) error
}
