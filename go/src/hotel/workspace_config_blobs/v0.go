package workspace_config_blobs

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
)

type V0 struct {
	Defaults repo_configs.DefaultsV1OmitEmpty `toml:"defaults,omitempty"`
	// FileExtensions file_extensions.V1    `toml:"file-extensions"`
	// PrintOptions   options_print.V0      `toml:"cli-output"`
	// Tools          options_tools.Options `toml:"tools"`

	Query string `toml:"query,omitempty"`
}

func (blob V0) GetDefaults() repo_configs.Defaults {
	return blob.Defaults
}

func (blob V0) GetDefaultQueryString() string {
	return blob.Query
}

type blobV0Coder struct{}

func (blobV0Coder) DecodeFrom(
	typedBlob TypedConfig,
	reader *bufio.Reader,
) (bytesRead int64, err error) {
	blob := Config(&V0{})
	typedBlob.Blob = &blob

	dec := toml.NewDecoder(reader)

	if err = dec.Decode(*typedBlob.Blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (blobV0Coder) EncodeTo(
	typedBlob TypedConfig,
	writer *bufio.Writer,
) (bytesWritten int64, err error) {
	dec := toml.NewEncoder(writer)

	if err = dec.Encode(*typedBlob.Blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
