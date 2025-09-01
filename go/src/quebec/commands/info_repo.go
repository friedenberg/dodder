package commands

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/age"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/golf/command"
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

		case "type":
			repo.GetUI().Print(configBlob.GetRepoType())

		case "id":
			repo.GetUI().Print(configBlob.GetRepoId())

			// TODO switch to `blob_stores.N.compression_type`
		case "compression-type":
			// TODO read default blob store and expose config
			repo.GetUI().Print(
				blobIOWrapper.GetBlobCompression(),
			)

			// TODO switch to `blob_stores.N.age_encryption`
			// TODO switch to encryption interface
		case "blob_stores-0-encryption":
			// TODO read default blob store and expose config
			for _, i := range blobIOWrapper.GetBlobEncryption().(*age.Age).Identities {
				repo.GetUI().Print(i)
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
			// TODO migrate this to config
			var pubKey markl.Id

			if err := markl.SetMerkleIdWithFormat(
				&pubKey,
				markl.FormatIdRepoPubKeyV1,
				configBlob.GetPublicKey(),
			); err != nil {
				repo.Cancel(err)
			}

			repo.GetUI().Print(pubKey.StringWithFormat())

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
