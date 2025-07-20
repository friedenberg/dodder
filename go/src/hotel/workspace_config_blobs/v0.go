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

func (blob V0) GetWorkspaceConfig() Blob {
	return blob
}

func (blob V0) GetDefaults() repo_configs.Defaults {
	return blob.Defaults
}

func (blob V0) GetDefaultQueryGroup() string {
	return blob.Query
}

type blobV0Coder struct{}

func (blobV0Coder) DecodeFrom(
	subject TypeWithBlob,
	reader *bufio.Reader,
) (n int64, err error) {
	blob := Blob(&V0{})
	subject.Blob = &blob

	dec := toml.NewDecoder(reader)

	if err = dec.Decode(*subject.Blob); err != nil {
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
	subject TypeWithBlob,
	writer *bufio.Writer,
) (n int64, err error) {
	dec := toml.NewEncoder(writer)

	if err = dec.Encode(*subject.Blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
