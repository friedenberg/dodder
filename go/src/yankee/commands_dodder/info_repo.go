package commands_dodder

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/golf/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/kilo/command_components_madder"
	"code.linenisgreat.com/dodder/go/src/xray/command_components_dodder"
)

func init() {
	// TODO rename to repo-info
	utility.AddCmd("info-repo", &InfoRepo{})
}

type InfoRepo struct {
	command_components_madder.BlobStoreConfig
	command_components_dodder.EnvRepo
}

func (cmd InfoRepo) Run(req command.Request) {
	args := req.PopArgs()
	env := cmd.MakeEnvRepo(req, false)

	// TODO should this be the private config flavor?
	configPublicTypedBlob := env.GetConfigPublic()
	configPublicBlob := configPublicTypedBlob.Blob

	configPrivateTypedBlob := env.GetConfigPrivate()
	configPrivateBlob := configPrivateTypedBlob.Blob

	// storeVersion := configPublicBlob.GetStoreVersion()
	defaultblobStore := env.GetDefaultBlobStore()
	blobIOWrapper := defaultblobStore.GetBlobIOWrapper()

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
			if _, err := genesis_configs.CoderPublic.EncodeTo(
				&configPublicTypedBlob,
				env.GetUIFile(),
			); err != nil {
				env.Cancel(err)
			}

		case "store-version":
			env.GetUI().Print(configPublicBlob.GetStoreVersion())

		case "id":
			env.GetUI().Print(configPublicBlob.GetRepoId())

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

		case "blob_stores-0-base-path":
			env.GetUI().Print(
				directory_layout.GetDefaultBlobStore(env).GetBase(),
			)

		case "blob_stores-0-config-path":
			env.GetUI().Print(
				directory_layout.GetDefaultBlobStore(env).GetConfig(),
			)

		case "blob_stores-0-config":
			blobStoreConfig := defaultblobStore.ConfigNamed.Config

			if err := cmd.PrintBlobStoreConfig(
				env,
				&blob_store_configs.TypedConfig{
					Type: blobStoreConfig.Type,
					Blob: blobStoreConfig.Blob,
				},
				env.GetUI().GetFile(),
			); err != nil {
				env.Cancel(err)
				return
			}

		// case "dir-blob_stores":
		// 	env.GetUI().Print(env.MakePathBlobStore())

		// TODO make dynamic and parse index
		// case "dir-blob_stores-0-blobs":
		// 	dir, target := directory_layout.GetBlobStoreConfigPath(
		// 		env,
		// 		0,
		// 		"default",
		// 	)

		// 	env.GetUI().Print(filepath.Join(dir, target))

		// TODO make dynamic and parse index
		// case "dir-blob_stores-0-inventory_lists":
		// 	if store_version.LessOrEqual(storeVersion, store_version.V10) {
		// 		env.GetUI().Print(
		// 			env.DirFirstBlobStoreInventoryLists(),
		// 		)
		// 	} else {
		// 		env.GetUI().Print(
		// 			env.DirFirstBlobStoreBlobs(),
		// 		)
		// 	}

		case "pubkey":
			env.GetUI().Print(
				configPublicBlob.GetPublicKey().StringWithFormat(),
			)

		case "seckey":
			env.Cancel(errors.Err405MethodNotAllowed)

			env.GetUI().Print(
				configPrivateBlob.GetPrivateKey().StringWithFormat(),
			)

		case "xdg":
			exdg := env.GetXDG()

			dotenv := xdg.Dotenv{
				XDG: &exdg,
			}

			if _, err := dotenv.WriteTo(env.GetUIFile()); err != nil {
				env.Cancel(err)
			}
		}
	}
}
