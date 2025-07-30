package remote_http

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/comments"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/india/log_remote_inventory_lists"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

func MakeClient(
	envUI env_ui.Env,
	transport http.RoundTripper,
	localRepo repo.LocalRepo,
	typedBlobStore inventory_list_coders.Closet,
) *client {
	client := &client{
		envUI: envUI,
		http: http.Client{
			Transport: transport,
		},
		localRepo:      localRepo,
		typedBlobStore: typedBlobStore,
	}

	client.Initialize()

	return client
}

type client struct {
	envUI           env_ui.Env
	configImmutable genesis_configs.TypedConfigPublic
	http            http.Client
	localRepo       repo.LocalRepo
	typedBlobStore  inventory_list_coders.Closet

	logRemoteInventoryLists log_remote_inventory_lists.Log
}

func (client *client) Initialize() {
	var request *http.Request

	{
		var err error

		if request, err = client.newRequest(
			"GET",
			"/config-immutable",
			nil,
		); err != nil {
			client.envUI.Cancel(err)
		}
	}

	var response *http.Response

	{
		var err error

		if response, err = client.http.Do(request); err != nil {
			errors.ContextCancelWithErrorAndFormat(
				client.envUI,
				err,
				"failed to read response",
			)
		}
	}

	if _, err := genesis_configs.CoderPublic.DecodeFrom(
		&client.configImmutable,
		response.Body,
	); err != nil {
		errors.ContextCancelWithErrorAndFormat(
			client.envUI,
			err,
			"failed to read remote immutable config",
		)
	}

	client.logRemoteInventoryLists = log_remote_inventory_lists.Make(
		client.localRepo.GetEnvRepo(),
		client.localRepo.GetEnvRepo(),
	)
}

func (client *client) GetEnv() env_ui.Env {
	return client.envUI
}

func (client *client) GetImmutableConfigPublic() genesis_configs.ConfigPublic {
	return client.configImmutable.Blob
}

func (client *client) GetImmutableConfigPublicType() ids.Type {
	return client.configImmutable.Type
}

func (client *client) GetInventoryListStore() sku.InventoryListStore {
	return client
}

func (client *client) GetTypedInventoryListBlobStore() inventory_list_coders.Closet {
	return client.typedBlobStore
}

func (client *client) GetObjectStore() sku.RepoStore {
	return nil
}

func (client *client) MakeImporter(
	options sku.ImporterOptions,
	storeOptions sku.StoreOptions,
) sku.Importer {
	panic(comments.Implement())
}

func (client *client) ImportList(
	list *sku.List,
	i sku.Importer,
) (err error) {
	return comments.Implement()
}

func (client *client) MakeExternalQueryGroup(
	builderOptions query.BuilderOption,
	externalQueryOptions sku.ExternalQueryOptions,
	args ...string,
) (qg *query.Query, err error) {
	err = comments.Implement()
	return
}

func (client *client) MakeInventoryList(
	queryGroup *query.Query,
) (list *sku.List, err error) {
	var request *http.Request
	listTypeString := client.GetImmutableConfigPublic().GetInventoryListTypeString()

	if request, err = client.newRequest(
		"GET",
		fmt.Sprintf("/query/%s/%s",
			url.QueryEscape(listTypeString),
			url.QueryEscape(queryGroup.String()),
		),
		nil,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var response *http.Response

	if response, err = client.http.Do(request); err != nil {
		err = errors.ErrorWithStackf("failed to read response: %w", err)
		return
	}

	if err = ReadErrorFromBodyOnNot(response, 200); err != nil {
		err = errors.Wrap(err)
		return
	}

	inventoryListCoderCloset := client.localRepo.GetTypedInventoryListBlobStore()

	if list, err = inventoryListCoderCloset.ReadInventoryListBlob(
		client.localRepo.GetEnvRepo(),
		ids.GetOrPanic(listTypeString).Type,
		bufio.NewReader(response.Body),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// func (remoteHTTP *HTTP) PullQueryGroupFromRemote2(
// 	remote repo.ReadWrite,
// 	options repo.RemoteTransferOptions,
// 	queryStrings ...string,
// ) (err error) {
// 	var qg *query.Group

// 	if qg, err = remoteHTTP.MakeQueryGroup(queryStrings...); err != nil {
// 		err = errors.Wrap(err)
// 		return
// 	}

// 	if err = remoteHTTP.PullQueryGroupFromRemote(
// 		remote,
// 		qg,
// 		options,
// 	); err != nil {
// 		err = errors.Wrap(err)
// 		return
// 	}

// 	return
// }

func (client *client) PullQueryGroupFromRemote(
	remote repo.Repo,
	queryGroup *query.Query,
	options repo.RemoteTransferOptions,
) (err error) {
	return client.pullQueryGroupFromWorkingCopy(
		remote.(repo.WorkingCopy),
		queryGroup,
		options,
	)
}

func (client *client) pullQueryGroupFromWorkingCopy(
	remote repo.WorkingCopy,
	queryGroup *query.Query,
	options repo.RemoteTransferOptions,
) (err error) {
	var list *sku.List

	if list, err = remote.MakeInventoryList(queryGroup); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO local / remote version negotiation

	listFormat := client.GetInventoryListStore().FormatForVersion(
		client.localRepo.GetImmutableConfigPublic().GetStoreVersion(),
	)

	buffer := bytes.NewBuffer(nil)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(buffer)
	defer repoolBufferedWriter()

	for {
		errors.ContextContinueOrPanic(client.envUI)

		// TODO make a reader version of inventory lists to avoid allocation
		if _, err = inventory_list_coders.WriteInventoryList(
			client.envUI,
			listFormat,
			quiter.MakeSeqErrorFromSeq(list.All()),
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		var response *http.Response

		if err = bufferedWriter.Flush(); err != nil {
			err = errors.Wrap(err)
			return
		}

		{
			var request *http.Request

			if request, err = client.newRequest(
				"POST",
				"/inventory_lists",
				buffer,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			if options.AllowMergeConflicts {
				// TODO move to constant
				request.Header.Add(
					"x-dodder-remote_transfer_options-allow_merge_conflicts",
					"true",
				)
			}

			if response, err = client.http.Do(request); err != nil {
				err = errors.ErrorWithStackf("failed to read response: %w", err)
				return
			}
		}

		if err = ReadErrorFromBodyOnNot(
			response,
			http.StatusCreated,
			http.StatusExpectationFailed,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		bufferedReader := bufio.NewReader(response.Body)

		var listMissingSkus *sku.List

		if listMissingSkus, err = client.typedBlobStore.ReadInventoryListBlob(
			client.GetEnv(),
			ids.GetOrPanic(
				client.configImmutable.Blob.GetInventoryListTypeString(),
			).Type,
			bufferedReader,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = response.Body.Close(); err != nil {
			err = errors.Wrap(err)
			return
		}

		ui.Log().Print(
			"received missing blob list: %d",
			listMissingSkus.Len(),
		)

		// if options.IncludeBlobs {
		for expected := range listMissingSkus.All() {
			errors.ContextContinueOrPanic(client.envUI)

			if err = client.WriteBlobToRemote(
				remote.GetBlobStore(),
				sha.MustWithDigester(expected.GetBlobId()),
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
		// }

		if response.StatusCode == http.StatusCreated {
			ui.Log().Print("done")
			return
		}

		buffer.Reset()
		bufferedWriter.Reset(buffer)
	}
}

func (client *client) ReadObjectHistory(
	oid *ids.ObjectId,
) (skus []*sku.Transacted, err error) {
	err = comments.Implement()
	return
}
