package commands

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/genesis_config_io"
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

	if len(args) == 0 {
		args = []string{"store-version"}
	}

	for _, arg := range args {
		switch strings.ToLower(arg) {
		default:
			repo.CancelWithBadRequestf("unsupported info key: %q", arg)

		case "config-immutable":
			if _, err := (genesis_config_io.CoderPublic{}).EncodeTo(
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

		case "compression-type":
			// TODO read default blob store and expose config
			repo.GetUI().Print(
				configBlob.GetBlobStoreConfigImmutable().GetBlobCompression(),
			)

		case "age-encryption":
			// TODO read default blob store and expose config
			for _, i := range configBlob.GetBlobStoreConfigImmutable().GetBlobEncryption().(*age.Age).Identities {
				repo.GetUI().Print(i)
			}

		case "dir.blob-stores.1.blobs":
			repo.GetUI().Print(
				repo.DirFirstBlobStoreBlobs(),
			)

		case "dir.blob-stores.1.inventory_lists":
			repo.GetUI().Print(
				repo.DirFirstBlobStoreInventoryLists(),
			)

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
