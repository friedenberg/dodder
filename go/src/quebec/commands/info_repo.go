package commands

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
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
	blobStoreConfigDefault := repo.GetDefaultBlobStore()

	if len(args) == 0 {
		args = []string{"store-version"}
	}

	for _, arg := range args {
		switch strings.ToLower(arg) {
		default:
			repo.CancelWithBadRequestf("unsupported info key: %q", arg)

		case "config-immutable":
			if _, err := genesis_config.CoderPublic.EncodeTo(
				&configTypedBlob,
				repo.GetUIFile(),
			); err != nil {
				repo.CancelWithError(err)
			}

		case "store-version":
			repo.GetUI().Print(configBlob.GetStoreVersion())

		case "type":
			repo.GetUI().Print(configBlob.GetRepoType())

		case "id":
			repo.GetUI().Print(configBlob.GetRepoId())

			// TODO switch to `blob_stores.N.compression_type`
		case "compression-type":
			// TODO read default blob store and expose config
			repo.GetUI().Print(
				blobStoreConfigDefault.GetBlobCompression(),
			)

			// TODO switch to `blob_stores.N.age_encryption`
			// TODO switch to encryption interface
		case "age-encryption":
			// TODO read default blob store and expose config
			for _, i := range blobStoreConfigDefault.GetBlobEncryption().(*age.Age).Identities {
				repo.GetUI().Print(i)
			}

			// TODO switch to `blob_stores`
		case "dir.blob-stores":
			repo.GetUI().Print(
				repo.DirBlobStores(),
			)

			// TODO switch to `blob_stores`
		case "dir.blob-stores.1.blobs":
			repo.GetUI().Print(
				repo.DirFirstBlobStoreBlobs(),
			)

			// TODO switch to `blob_stores`
		case "dir.blob-stores.1.inventory_lists":
			if store_version.LessOrEqual(storeVersion, store_version.V10) {
				repo.GetUI().Print(
					repo.DirFirstBlobStoreInventoryLists(),
				)
			} else {
				repo.GetUI().Print(
					repo.DirFirstBlobStoreBlobs(),
				)
			}

		case "xdg":
			ecksDeeGee := repo.GetXDG()

			dotenv := xdg.Dotenv{
				XDG: &ecksDeeGee,
			}

			if _, err := dotenv.WriteTo(repo.GetUIFile()); err != nil {
				repo.CancelWithError(err)
			}
		}
	}
}
