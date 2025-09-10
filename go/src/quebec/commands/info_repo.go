package commands

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	// TODO rename to repo-info
	command.Register("info-repo", &InfoRepo{})
}

type InfoRepo struct {
	command_components.EnvRepo
}

func (cmd InfoRepo) Run(req command.Request) {
	args := req.PopArgs()
	repo := cmd.MakeEnvRepo(req, false)

	// TODO should this be the private config flavor?
	configTypedBlob := repo.GetConfigPublic()
	configBlob := configTypedBlob.Blob
	storeVersion := configBlob.GetStoreVersion()
	blobStore := repo.GetDefaultBlobStore()
	blobIOWrapper := blobStore.GetBlobIOWrapper()

	if len(args) == 0 {
		args = []string{"store-version"}
	}

	for _, arg := range args {
		switch strings.ToLower(arg) {
		default:
			errors.ContextCancelWithBadRequestf(
				repo,
				"unsupported info key: %q",
				arg,
			)

		case "config-immutable":
			if _, err := genesis_configs.CoderPublic.EncodeTo(
				&configTypedBlob,
				repo.GetUIFile(),
			); err != nil {
				repo.Cancel(err)
			}

		case "store-version":
			repo.GetUI().Print(configBlob.GetStoreVersion())

		case "id":
			repo.GetUI().Print(configBlob.GetRepoId())

			// TODO switch to `blob_stores.N.compression_type`
		case "compression-type":
			// TODO read default blob store and expose config
			repo.GetUI().Print(
				blobIOWrapper.GetBlobCompression(),
			)

			// TODO switch to `blob_stores.N.age_encryption`
		case "blob_stores-0-encryption":
			repo.GetUI().Print(
				blobIOWrapper.GetBlobEncryption().StringWithFormat(),
			)

		case "blob_stores-0-config-path":
			repo.GetUI().Print(
				repo.DirBlobStoreConfigs(
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

			if configLocalMutable, ok := blobStoreConfig.Blob.(blob_store_configs.ConfigLocalMutable); ok {
				configLocalMutable.SetBasePath(
					blobStore.BlobStoreConfigNamed.BasePath,
				)
			}

			if _, err := blob_store_configs.Coder.EncodeTo(
				&blob_store_configs.TypedConfig{
					Type: blobStoreConfig.Type,
					Blob: blobStoreConfig.Blob,
				},
				repo.GetUI().GetFile(),
			); err != nil {
				repo.Cancel(err)
				return
			}

		case "dir-blob_stores":
			repo.GetUI().Print(
				repo.DirBlobStores(),
			)

			// TODO make dynamic and parse index
		case "dir-blob_stores-0-blobs":
			repo.GetUI().Print(
				repo.DirFirstBlobStoreBlobs(),
			)

			// TODO make dynamic and parse index
		case "dir-blob_stores-0-inventory_lists":
			if store_version.LessOrEqual(storeVersion, store_version.V10) {
				repo.GetUI().Print(
					repo.DirFirstBlobStoreInventoryLists(),
				)
			} else {
				repo.GetUI().Print(
					repo.DirFirstBlobStoreBlobs(),
				)
			}

		case "pubkey":
			repo.GetUI().Print(configBlob.GetPublicKey().StringWithFormat())

		case "xdg":
			ecksDeeGee := repo.GetXDG()

			dotenv := xdg.Dotenv{
				XDG: &ecksDeeGee,
			}

			if _, err := dotenv.WriteTo(repo.GetUIFile()); err != nil {
				repo.Cancel(err)
			}
		}
	}
}
