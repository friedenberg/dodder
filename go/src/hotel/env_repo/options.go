package env_repo

import (
	"os"
)

type Options struct {
	BasePath                string
	PermitNoDodderDirectory bool
	MakeXDGDirectories      bool
}

func (o Options) GetReadOnlyBlobStorePath() string {
	return os.Getenv("DODDER_READ_ONLY_BLOB_STORE_PATH")
}
