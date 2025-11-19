package organize_text

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/tag_paths"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
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
		return out, err
	}

	return out, err
}

// TODO: claude: refactor to use single call to `GetMetadataMutable()` at start
// of loops to simplify subsequent nested method calls
func (assignment *Assignment) addToSet(
	ot *Text,
	output SkuMapWithOrder,
	objectsFromBefore sku.SkuTypeSet,
) (err error) {
	expanded := ids.MakeTagMutableSet()

	if err = assignment.AllTags(expanded); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for _, organizeObject := range assignment.All() {
		var outputObject sku.SkuType

		objectKey := keyer.GetKey(organizeObject.sku)

		previouslyProcessedObject, wasPreviouslyProcessed := output.m[objectKey]

		if !wasPreviouslyProcessed {
			outputObject = ot.ObjectFactory.Get()

			ot.ObjectFactory.ResetWith(outputObject, organizeObject.sku)

			if !ot.Metadata.Type.IsEmpty() {
				outputObject.GetSkuExternal().GetMetadataMutable().GetTypeMutable().ResetWith(
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

				outputObject.GetSkuExternal().GetMetadataMutable().GetTypeMutable().ResetWith(
					objectOriginal.GetSkuExternal().GetMetadata().GetType(),
				)

				outputObject.GetSkuExternal().GetSkuExternal().Metadata.GetBlobDigestMutable().ResetWithMarklId(
					objectOriginal.GetSkuExternal().GetSkuExternal().Metadata.GetBlobDigest(),
				)

				outputObject.GetSkuExternal().GetSkuExternal().GetMetadataMutable().GetTypeMutable().ResetWith(
					objectOriginal.GetSkuExternal().GetSkuExternal().GetMetadata().GetType(),
				)

				outputObject.SetState(objectOriginal.GetState())

				{
					src := &objectOriginal.GetSkuExternal().Metadata
					dst := &outputObject.GetSkuExternal().Metadata
					dst.Fields = objectOriginal.GetSkuExternal().Metadata.Fields[:0]
					dst.Fields = append(dst.Fields, src.Fields...)
				}
			}

			outputMetadata := outputObject.GetSkuExternal().GetMetadataMutable()

			for tag := range ot.Metadata.AllPtr() {
				if organizeObject.tipe == tag_paths.TypeUnknown {
					continue
				}

				if _, ok := outputMetadata.GetIndex().GetTagPaths().All.ContainsString(
					tag.String(),
				); ok {
					continue
				}

				outputObject.GetSkuExternal().AddTagPtr(tag)
			}

			if !ot.Metadata.Type.IsEmpty() {
				outputObject.GetSkuExternal().GetMetadataMutable().GetTypeMutable().ResetWith(
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

		if err = outputObject.GetSkuExternal().GetMetadataMutable().GetDescriptionMutable().Set(
			organizeObject.GetSkuExternal().GetMetadata().GetDescription().String(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if !organizeObject.GetSkuExternal().GetMetadata().GetType().IsEmpty() {
			if err = outputObject.GetSkuExternal().GetMetadataMutable().GetTypeMutable().Set(
				organizeObject.GetSkuExternal().GetMetadata().GetType().String(),
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		if !organizeObject.tipe.IsDirectOrSelf() {
			return err
		}

		quiter.AppendSeq(
			outputObject.GetSkuExternal().GetMetadataMutable().GetCommentsMutable(),
			organizeObject.GetSkuExternal().GetMetadataMutable().GetComments(),
		)

		for e := range organizeObject.GetSkuExternal().Metadata.GetTags().AllPtr() {
			if err = outputObject.GetSkuExternal().AddTagPtr(e); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		for e := range expanded.AllPtr() {
			outputObject.GetSkuExternal().AddTagPtr(e)
		}
	}

	for _, c := range assignment.Children {
		if err = c.addToSet(ot, output, objectsFromBefore); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
