package commands

import (
	"encoding/json"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("read-blob", &ReadBlob{})
}

type ReadBlob struct {
	command_components.EnvRepo
}

type readBlobEntry struct {
	Blob string `json:"blob"`
}

func (c ReadBlob) Run(dep command.Request) {
	repoLayout := c.MakeEnvRepo(dep, false)

	dec := json.NewDecoder(repoLayout.GetInFile())

	for {
		var entry readBlobEntry

		if err := dec.Decode(&entry); err != nil {
			if errors.IsEOF(err) {
				err = nil
			} else {
				repoLayout.Cancel(err)
			}

			return
		}

		{
			var err error

			if _, err = c.readOneBlob(repoLayout, entry); err != nil {
				repoLayout.Cancel(err)
			}
		}
	}
}

func (ReadBlob) readOneBlob(
	envRepo env_repo.Env,
	entry readBlobEntry,
) (digest interfaces.BlobId, err error) {
	var writeCloser interfaces.WriteCloseBlobIdGetter

	if writeCloser, err = envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if _, err = io.Copy(writeCloser, strings.NewReader(entry.Blob)); err != nil {
		err = errors.Wrap(err)
		return
	}

	digest = writeCloser.GetBlobId()

	return
}
