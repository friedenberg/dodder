package env_dir

import (
	"io"
	"os"
)

// TODO decouple Config from all of these

type ReadOptions struct {
	Config
	*os.File
}

type FileReadOptions struct {
	Config
	Path string
}

type WriteOptions struct {
	io.Writer
}

type MoveOptions struct {
	TemporaryFS
	ErrorOnAttemptedOverwrite bool
	FinalPath                 string
	GenerateFinalPathFromSha  bool
}
