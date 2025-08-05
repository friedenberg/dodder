package workspace_config_blobs

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
)

type (
	Config interface {
		GetDefaults() repo_configs.Defaults
	}

	ConfigTemporary interface {
		Config
		temporaryWorkspace()
	}

	ConfigWithDefaultQueryString interface {
		Config
		GetDefaultQueryString() string
	}
)

var (
	_ ConfigWithDefaultQueryString = V0{}
	_ ConfigTemporary              = Temporary{}
)

type TypedConfig = *triple_hyphen_io.TypedBlob[*Config]

var coders = map[string]interfaces.CoderBufferedReadWriter[TypedConfig]{
	ids.TypeTomlWorkspaceConfigV0: blobV0Coder{},
}

var Coder = triple_hyphen_io.Coder[TypedConfig]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[*Config]{},
	Blob:     triple_hyphen_io.CoderTypeMap[*Config](coders),
}

func DecodeFromFile(
	object TypedConfig,
	path string,
) (err error) {
	var file *os.File

	if file, err = files.OpenExclusiveReadOnly(path); err != nil {
		if !errors.IsNotExist(err) {
			err = errors.Wrap(err)
		}

		return
	}

	defer errors.DeferredCloser(&err, file)

	if _, err = Coder.DecodeFrom(object, file); err != nil {
		err = errors.Wrapf(err, "File: %q", file.Name())
		return
	}

	return
}

func EncodeToFile(
	object TypedConfig,
	path string,
) (err error) {
	var file *os.File

	if file, err = files.CreateExclusiveWriteOnly(path); err != nil {
		if !errors.IsNotExist(err) {
			err = errors.Wrap(err)
		}

		return
	}

	defer errors.DeferredCloser(&err, file)

	if _, err = Coder.EncodeTo(object, file); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
