package repo_blobs

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/remote_connection_types"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

type (
	Blob interface {
		GetPublicKey() interfaces.MarklId
		IsRemote() bool
	}

	BlobMutable interface {
		Blob
		SetPublicKey(interfaces.MarklId)
	}

	BlobXDG interface {
		Blob
		MakeXDG(utilityName string) xdg.XDG
	}

	BlobOverridePath interface {
		Blob
		GetOverridePath() string
	}

	BlobUri interface {
		Blob
		GetUri() values.Uri
	}
)

func GetSupportedConnectionTypes(
	blob Blob,
) interfaces.SetLike[remote_connection_types.Type] {
	if blob.IsRemote() {
		return collections_value.MakeValueSetValue(
			nil,
			remote_connection_types.TypeSocketUnix,
			remote_connection_types.TypeUrl,
			remote_connection_types.TypeStdioSSH,
		)
	} else {
		return collections_value.MakeValueSetValue(
			nil,
			remote_connection_types.TypeNative,
			remote_connection_types.TypeNativeLocalOverridePath,
			remote_connection_types.TypeSocketUnix,
			remote_connection_types.TypeStdioLocal,
		)
	}
}
