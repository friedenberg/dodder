package commands_madder

import (
	"encoding/json"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	utility.AddCmd("read", &Read{})
}

type Read struct {
	command_components_madder.EnvRepo
}

type readBlobEntry struct {
	Blob string `json:"blob"`
}

func (cmd Read) Run(dep command.Request) {
	envRepo := cmd.MakeEnvRepo(dep, false)

	decoder := json.NewDecoder(envRepo.GetInFile())

	for {
		var entry readBlobEntry

		if err := decoder.Decode(&entry); err != nil {
			if errors.IsEOF(err) {
				err = nil
			} else {
				envRepo.Cancel(err)
			}

			return
		}

		{
			var err error

			if _, err = cmd.readOneBlob(envRepo, entry); err != nil {
				envRepo.Cancel(err)
			}
		}
	}
}

func (Read) readOneBlob(
	envRepo env_repo.Env,
	entry readBlobEntry,
) (digest interfaces.MarklId, err error) {
	var writeCloser interfaces.BlobWriter

	if writeCloser, err = envRepo.GetDefaultBlobStore().MakeBlobWriter(nil); err != nil {
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
