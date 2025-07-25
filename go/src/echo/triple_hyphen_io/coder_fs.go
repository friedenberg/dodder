package triple_hyphen_io

import (
	"bytes"
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
) (typedBlob TypedBlob[BLOB], err error) {
	var file *os.File

	if file, err = files.OpenExclusiveReadOnly(
		path,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	if _, err = coders.DecodeFrom(&typedBlob, file); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// For decodes where typedBlob.Blob should be populated with an empty struct
// when the file is missing
func DecodeFromFileOrEmptyBuffer[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
](
	coders CoderToTypedBlob[BLOB],
	path string,
	permitNotExist bool,
) (typedBlob TypedBlob[BLOB], err error) {
	typedBlob, err = DecodeFromFile(coders, path)

	if err == nil {
		return
	}

	if _, err = coders.DecodeFrom(&typedBlob, bytes.NewBuffer(nil)); err != nil {
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
