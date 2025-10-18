package commands_madder

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	// TODO rename to repo-info
	utility.AddCmd("info-repo", &InfoRepo{})
}

type InfoRepo struct {
	command_components_madder.EnvBlobStore
}

func (cmd InfoRepo) Run(req command.Request) {
	args := req.PopArgs()
	env := cmd.MakeEnvBlobStore(req)

	blobStore := env.GetDefaultBlobStore()
	blobStoreConfig := blobStore.Config
	blobIOWrapper := blobStore.GetBlobIOWrapper()

	if len(args) == 0 {
		args = []string{"store-version"}
	}

	for _, arg := range args {
		switch strings.ToLower(arg) {
		default:
			errors.ContextCancelWithBadRequestf(
				env,
				"unsupported info key: %q",
				arg,
			)

		case "config-immutable":
			if _, err := blob_store_configs.Coder.EncodeTo(
				&blobStoreConfig,
				env.GetUIFile(),
			); err != nil {
				env.Cancel(err)
			}

		case "name":
			// TODO

			// TODO switch to `blob_stores.N.compression_type`
		case "compression-type":
			// TODO read default blob store and expose config
			env.GetUI().Print(
				blobIOWrapper.GetBlobCompression(),
			)

			// TODO switch to `blob_stores.N.age_encryption`
		case "blob_stores-0-encryption":
			env.GetUI().Print(
				blobIOWrapper.GetBlobEncryption().StringWithFormat(),
			)

		case "blob_stores-0-config-path":
			env.GetUI().Print(
				env.DirBlobStoreConfigs(
					fmt.Sprintf(
						"%d-default.%s",
						0,
						env_repo.FileNameBlobStoreConfig,
					),
				),
			)

		case "blob_stores-0-config":
			// TODO this is gross, fix it
			blobStoreConfig := blobStore.BlobStoreConfigNamed.Config

			if _, err := blob_store_configs.Coder.EncodeTo(
				&blob_store_configs.TypedConfig{
					Type: blobStoreConfig.Type,
					Blob: blobStoreConfig.Blob,
				},
				env.GetUI().GetFile(),
			); err != nil {
				env.Cancel(err)
				return
			}

		case "dir-blob_stores":
			env.GetUI().Print(env.MakePathBlobStore())

			// TODO make dynamic and parse index
		case "dir-blob_stores-0-blobs":
			env.GetUI().Print(
				env.DirFirstBlobStoreBlobs(),
			)

		case "xdg":
			ecksDeeGee := env.GetXDG()

			dotenv := xdg.Dotenv{
				XDG: &ecksDeeGee,
			}

			if _, err := dotenv.WriteTo(env.GetUIFile()); err != nil {
				env.Cancel(err)
			}
		}
	}
}
