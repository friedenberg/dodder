package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
)

type ConfigNamed struct {
	Path   directory_layout.BlobStorePath
	Config TypedConfig
}
