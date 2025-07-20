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
	Blob interface {
		GetDefaults() repo_configs.Defaults
		GetDefaultQueryGroup() string
	}
)

type TypeWithBlob = *triple_hyphen_io.TypedBlob[*Blob]

var typedCoders = map[string]interfaces.CoderBufferedReadWriter[TypeWithBlob]{
	ids.TypeTomlWorkspaceConfigV0: blobV0Coder{},
}

var Coder = triple_hyphen_io.Coder[TypeWithBlob]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[*Blob]{},
	Blob:     triple_hyphen_io.CoderTypeMap[*Blob](typedCoders),
}

func DecodeFromFile(
	object TypeWithBlob,
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
	object TypeWithBlob,
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
