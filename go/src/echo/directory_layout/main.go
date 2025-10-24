package directory_layout

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type DirectoryLayout interface {
	interfaces.RepoDirectoryLayout
	Initialize(xdg.XDG) error
}

func MakeDirectoryLayout(storeVersion store_version.Version) DirectoryLayout {
	if storeVersion.LessOrEqual(store_version.V10) {
		return &V1{}
	} else {
		return &V2{}
	}
}
