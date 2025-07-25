package local_working_copy

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_fmt"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

type (
	FormatTypeFuncConstructor func(
		*Repo,
		typed_blob_store.Type,
		interfaces.WriterAndStringWriter,
	) interfaces.FuncIter[*sku.Transacted]

	FormatTypeFuncConstructorEntry struct {
		Name        string
		description string
		FormatTypeFuncConstructor
	}
)

func makeFormatEntryFromTypeFormat(
	typeEntry FormatTypeFuncConstructorEntry,
) FormatFuncConstructorEntry {
	return FormatFuncConstructorEntry{
		description: typeEntry.description,
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return typeEntry.FormatTypeFuncConstructor(
				repo,
				repo.GetStore().GetTypedBlobStore().Type,
				writer,
			)
		},
	}
}

var typeFormatters = map[string]FormatTypeFuncConstructorEntry{
	"vim-syntax-type": {
		FormatTypeFuncConstructor: func(
			repo *Repo,
			typeBlobStore typed_blob_store.Type,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				var typeObject *sku.Transacted

				if typeObject, err = repo.GetStore().ReadTransactedFromObjectId(object.GetType()); err != nil {
					if collections.IsErrNotFound(err) {
						err = nil
					} else {
						err = errors.Wrap(err)
						return
					}
				}

				if typeObject == nil || typeObject.ObjectId.IsEmpty() ||
					typeObject.GetBlobSha().IsNull() {
					ty := ""

					switch object.GetGenre() {
					case genres.Type, genres.Tag, genres.Repo, genres.Config:
						ty = "toml"

					default:
						// TODO zettel default typ
					}

					if _, err = fmt.Fprintln(writer, ty); err != nil {
						err = errors.Wrap(err)
						return
					}

					return
				}

				var typeBlob type_blobs.Blob

				if typeBlob, _, err = typeBlobStore.ParseTypedBlob(
					typeObject.GetType(),
					typeObject.GetBlobSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer typeBlobStore.PutTypedBlob(typeObject.GetType(), typeBlob)

				if _, err = fmt.Fprintln(
					writer,
					typeBlob.GetVimSyntaxType(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"formatters": {
		FormatTypeFuncConstructor: func(
			repo *Repo,
			typeBlobStore typed_blob_store.Type,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(o *sku.Transacted) (err error) {
				var tt *sku.Transacted

				if tt, err = repo.GetStore().ReadTransactedFromObjectId(o.GetType()); err != nil {
					err = errors.Wrap(err)
					return
				}

				var ta type_blobs.Blob

				if ta, _, err = typeBlobStore.ParseTypedBlob(
					tt.GetType(),
					tt.GetBlobSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer typeBlobStore.PutTypedBlob(tt.GetType(), ta)

				lw := format.MakeLineWriter()

				for fn, f := range ta.GetFormatters() {
					fe := f.FileExtension

					if fe == "" {
						fe = fn
					}

					lw.WriteFormat("%s\t%s", fn, fe)
				}

				if _, err = lw.WriteTo(writer); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"formatter-uti-groups": {
		FormatTypeFuncConstructor: func(
			repo *Repo,
			typeBlobStore typed_blob_store.Type,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			fo := sku_fmt.MakeFormatterTypFormatterUTIGroups(
				repo.GetStore(),
				typeBlobStore,
			)

			return func(o *sku.Transacted) (err error) {
				if _, err = fo.Format(writer, o); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"hooks.on_pre_commit": {
		FormatTypeFuncConstructor: func(
			repo *Repo,
			typeBlobStore typed_blob_store.Type,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(o *sku.Transacted) (err error) {
				var blob type_blobs.Blob

				if blob, _, err = typeBlobStore.ParseTypedBlob(
					o.GetType(),
					o.GetBlobSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer typeBlobStore.PutTypedBlob(o.GetType(), blob)

				script := blob.GetStringLuaHooks()

				if script == "" {
					return
				}

				// TODO switch to typed variant
				var vp sku.LuaVMPoolV1

				if vp, err = repo.GetStore().MakeLuaVMPoolV1(o, script); err != nil {
					err = errors.Wrap(err)
					return
				}

				var vm *sku.LuaVMV1

				if vm, err = vp.Get(); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer vp.Put(vm)

				f := vm.GetField(vm.Top, "on_pre_commit")

				repo.GetUI().Print(f.String())

				return
			}
		},
	},
}
