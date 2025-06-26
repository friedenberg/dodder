package env_dir

import (
	"io"
	"os"
)

type ReadOptions struct {
	Config
	*os.File
}

type FileReadOptions struct {
	Config
	Path string
}

type WriteOptions struct {
	Config
	io.Writer
}

type MoveOptions struct {
	Config
	TemporaryFS
	ErrorOnAttemptedOverwrite bool
	FinalPath                 string
	GenerateFinalPathFromSha  bool
}
