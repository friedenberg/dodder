package organize_text

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

var keyer = sku.GetExternalLikeKeyer[sku.SkuType]()

func (ot *Text) GetSkus(
	original sku.SkuTypeSet,
) (out SkuMapWithOrder, err error) {
	out = MakeSkuMapWithOrder(original.Len())

	if err = ot.addToSet(
		ot,
		out,
		original,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a *Assignment) addToSet(
	ot *Text,
	output SkuMapWithOrder,
	objectsFromBefore sku.SkuTypeSet,
) (err error) {
	expanded := ids.MakeTagMutableSet()

	if err = a.AllTags(expanded); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, organizeObject := range a.All() {
		var outputObject sku.SkuType

		objectKey := keyer.GetKey(organizeObject.sku)

		previouslyProcessedObject, wasPreviouslyProcessed := output.m[objectKey]

		if !wasPreviouslyProcessed {
			outputObject = ot.ObjectFactory.Get()

			ot.ObjectFactory.ResetWith(outputObject, organizeObject.sku)

			if !ot.Metadata.Type.IsEmpty() {
				outputObject.GetSkuExternal().Metadata.Type.ResetWith(
					ot.Metadata.Type,
				)
			}

			outputObject.GetSkuExternal().RepoId.ResetWith(ot.Metadata.RepoId)

			output.Add(outputObject)

			objectOriginal, hasOriginal := objectsFromBefore.Get(objectKey)

			if hasOriginal {
				outputObject.GetSkuExternal().Metadata.GetBlobDigestMutable().ResetWithMarklId(
					objectOriginal.GetSkuExternal().Metadata.GetBlobDigest(),
				)

				outputObject.GetSkuExternal().Metadata.Type.ResetWith(
					objectOriginal.GetSkuExternal().Metadata.Type,
				)

				outputObject.GetSkuExternal().GetSkuExternal().Metadata.GetBlobDigestMutable().ResetWithMarklId(
					objectOriginal.GetSkuExternal().GetSkuExternal().Metadata.GetBlobDigest(),
				)

				outputObject.GetSkuExternal().GetSkuExternal().Metadata.Type.ResetWith(
					objectOriginal.GetSkuExternal().GetSkuExternal().Metadata.Type,
				)

				outputObject.SetState(objectOriginal.GetState())

				{
					src := &objectOriginal.GetSkuExternal().Metadata
					dst := &outputObject.GetSkuExternal().Metadata
					dst.Fields = objectOriginal.GetSkuExternal().Metadata.Fields[:0]
					dst.Fields = append(dst.Fields, src.Fields...)
				}
			}

			outputMetadata := outputObject.GetSkuExternal().GetMetadata()

			for e := range ot.Metadata.AllPtr() {
				if organizeObject.tipe == tag_paths.TypeUnknown {
					continue
				}

				if _, ok := outputMetadata.Cache.TagPaths.All.ContainsString(
					e.String(),
				); ok {
					continue
				}

				outputObject.GetSkuExternal().AddTagPtr(e)
			}

			if !ot.Metadata.Type.IsEmpty() {
				outputObject.GetSkuExternal().Metadata.Type.ResetWith(
					ot.Metadata.Type,
				)
			}
		} else {
			outputObject = previouslyProcessedObject.sku
		}

		if organizeObject.GetSkuExternal().ObjectId.String() == "" {
			panic(fmt.Sprintf("%s: object id is nil", organizeObject))
		}

		if outputObject == nil {
			panic("empty object")
		}

		if err = outputObject.GetSkuExternal().Metadata.Description.Set(
			organizeObject.GetSkuExternal().Metadata.Description.String(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if !organizeObject.GetSkuExternal().Metadata.Type.IsEmpty() {
			if err = outputObject.GetSkuExternal().Metadata.Type.Set(
				organizeObject.GetSkuExternal().Metadata.Type.String(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		if !organizeObject.tipe.IsDirectOrSelf() {
			return
		}

		outputObject.GetSkuExternal().Metadata.Comments = append(
			outputObject.GetSkuExternal().Metadata.Comments,
			organizeObject.GetSkuExternal().Metadata.Comments...,
		)

		for e := range organizeObject.GetSkuExternal().Metadata.GetTags().AllPtr() {
			if err = outputObject.GetSkuExternal().AddTagPtr(e); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		for e := range expanded.AllPtr() {
			outputObject.GetSkuExternal().AddTagPtr(e)
		}
	}

	for _, c := range a.Children {
		if err = c.addToSet(ot, output, objectsFromBefore); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
