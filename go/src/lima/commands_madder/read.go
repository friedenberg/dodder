package commands_madder

import (
	"encoding/json"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/juliett/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/command_components_madder"
)

func init() {
	utility.AddCmd("read", &Read{})
}

type Read struct {
	command_components_madder.EnvBlobStore
}

type readBlobEntry struct {
	Blob string `json:"blob"`
}

func (cmd Read) Run(dep command.Request) {
	envBlobStore := cmd.MakeEnvBlobStore(dep)

	decoder := json.NewDecoder(envBlobStore.GetInFile())

	for {
		var entry readBlobEntry

		if err := decoder.Decode(&entry); err != nil {
			if errors.IsEOF(err) {
				err = nil
			} else {
				envBlobStore.Cancel(err)
			}

			return
		}

		{
			var err error

			if _, err = cmd.readOneBlob(envBlobStore, entry); err != nil {
				envBlobStore.Cancel(err)
			}
		}
	}
}

func (Read) readOneBlob(
	envBlobStore env_repo.BlobStoreEnv,
	entry readBlobEntry,
) (digest interfaces.MarklId, err error) {
	var writeCloser interfaces.BlobWriter

	if writeCloser, err = envBlobStore.GetDefaultBlobStore().MakeBlobWriter(
		nil,
	); err != nil {
		err = errors.Wrap(err)
		return digest, err
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if _, err = io.Copy(writeCloser, strings.NewReader(entry.Blob)); err != nil {
		err = errors.Wrap(err)
		return digest, err
	}

	digest = writeCloser.GetMarklId()

	return digest, err
}
