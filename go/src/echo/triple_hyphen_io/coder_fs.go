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
	coders CoderToTypedBlob[BLOB],
	path string,
	permitNotExist bool,
) (typedBlob TypedBlob[BLOB], err error) {
	var reader io.Reader

	{
		var file *os.File

		if file, err = files.OpenExclusiveReadOnly(
			path,
		); err != nil {
			if errors.IsNotExist(err) && permitNotExist {
				err = nil
				reader = bytes.NewBuffer(nil)
			} else {
				err = errors.Wrap(err)
				return
			}
		} else {
			reader = file
			defer errors.DeferredCloser(&err, file)
		}
	}

	if _, err = coders.DecodeFrom(&typedBlob, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func EncodeToFile[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
](
	coders CoderToTypedBlob[BLOB],
	typedBlob *TypedBlob[BLOB],
	path string,
) (err error) {
	var file *os.File

	if file, err = files.CreateExclusiveWriteOnly(
		path,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	if _, err = coders.EncodeTo(typedBlob, file); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
