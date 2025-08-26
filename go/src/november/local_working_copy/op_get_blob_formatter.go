package local_working_copy

import (
	"maps"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// TODO add support for checked out types
func (local *Repo) GetBlobFormatter(
	tipe ids.Type,
	formatId string,
	utiGroup string,
) (blobFormatter script_config.RemoteScript, err error) {
	if tipe.GetType().IsEmpty() {
		err = errors.ErrorWithStackf("empty type")
		return
	}

	var typeObject *sku.Transacted

	if typeObject, err = local.GetStore().ReadTransactedFromObjectId(
		tipe.GetType(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var typeBlob type_blobs.Blob

	if typeBlob, _, err = local.GetStore().GetTypedBlobStore().Type.ParseTypedBlob(
		typeObject.GetType(),
		typeObject.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer local.GetStore().GetTypedBlobStore().Type.PutTypedBlob(
		typeObject.GetType(),
		typeBlob,
	)

	ok := false

	if utiGroup == "" {
		getBlobFormatter := func(formatId string) script_config.RemoteScript {
			var formatIds []string

			if formatId == "" {
				formatIds = []string{"text-edit", "text"}
			} else {
				formatIds = []string{formatId}
			}

			for _, formatId := range formatIds {
				blobFormatter, ok = typeBlob.GetFormatters()[formatId]

				if ok {
					return blobFormatter
				}
			}

			return nil
		}

		blobFormatter = getBlobFormatter(formatId)

		return
	}

	var g type_blobs.UTIGroup
	g, ok = typeBlob.GetFormatterUTIGroups()[utiGroup]

	if !ok {
		err = errors.BadRequestf(
			"no uti group: %q. Available groups: %s",
			utiGroup,
			slices.Collect(maps.Keys(typeBlob.GetFormatterUTIGroups())),
		)
		return
	}

	ft, ok := g.Map()[formatId]

	if !ok {
		err = errors.ErrorWithStackf(
			"no format id %q for uti group %q. Available groups: %s",
			formatId,
			utiGroup,
			slices.Collect(maps.Keys(g.Map())),
		)

		return
	}

	formatId = ft

	blobFormatter, ok = typeBlob.GetFormatters()[formatId]

	if !ok {
		ui.Err().Print("no matching format id")
		blobFormatter = nil
		// TODO-P2 allow option to error on missing format
		// err = errors.Normalf("no format id %q", actualFormatId)
		// return

		return
	}

	return
}
