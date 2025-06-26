package env_repo

import (
	"os"
)

type Options struct {
	BasePath             string
	PermitNoZitDirectory bool
	MakeXDGDirectories   bool
}

func (o Options) GetReadOnlyBlobStorePath() string {
	return os.Getenv("ZIT_READ_ONLY_BLOB_STORE_PATH")
}
