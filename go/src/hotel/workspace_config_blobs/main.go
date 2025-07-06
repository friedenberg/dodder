package workspace_config_blobs

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/golf/config_mutable_blobs"
)

const (
	TypeV0 = builtin_types.WorkspaceConfigTypeTomlV0
)

type (
	Blob interface {
		GetDefaults() config_mutable_blobs.Defaults
		GetDefaultQueryGroup() string
	}
)

type TypeWithBlob = *triple_hyphen_io2.TypedBlob[*Blob]

var typedCoders = map[string]interfaces.CoderBufferedReadWriter[TypeWithBlob]{
	TypeV0: blobV0Coder{},
}

var Coder = triple_hyphen_io2.Coder[TypeWithBlob]{
	Metadata: triple_hyphen_io2.TypedMetadataCoder[*Blob]{},
	Blob:     triple_hyphen_io2.CoderTypeMap[*Blob](typedCoders),
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
