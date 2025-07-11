package triple_hyphen_io

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

func DecodeFromFile[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
](
	ctx errors.Context,
	coders CoderToTypedBlob[BLOB],
	path string,
	permitNotExist bool,
) (typedBlob TypedBlob[BLOB]) {
	var reader io.Reader

	{
		var file *os.File
		var err error

		if file, err = files.OpenExclusiveReadOnly(
			path,
		); err != nil {
			if errors.IsNotExist(err) {
				err = nil
				reader = bytes.NewBuffer(nil)
			} else {
				ctx.CancelWithError(err)
				return
			}
		} else {
			reader = file
			defer ctx.MustClose(file)
		}
	}

	if _, err := coders.DecodeFrom(&typedBlob, reader); err != nil {
		ctx.CancelWithError(err)
	}

	return
}

func EncodeToFile[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
](
	ctx errors.Context,
	coders CoderToTypedBlob[BLOB],
	typedBlob *TypedBlob[BLOB],
	path string,
) {
	var file *os.File

	{
		var err error

		if file, err = files.CreateExclusiveWriteOnly(
			path,
		); err != nil {
			ctx.CancelWithError(err)
		}

		defer ctx.MustClose(file)
	}

	if _, err := coders.EncodeTo(typedBlob, file); err != nil {
		ctx.CancelWithError(err)
	}
}
