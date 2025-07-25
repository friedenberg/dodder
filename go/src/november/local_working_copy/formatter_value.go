package local_working_copy

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/chrest/go/src/bravo/client"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_io"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_fmt"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

// TODO switch to using fd.Std
func (local *Repo) MakeFormatFunc(
	format string,
	writer interfaces.WriterAndStringWriter,
) (output interfaces.FuncIter[*sku.Transacted], err error) {
	if writer == nil {
		writer = local.GetUIFile()
	}

	if after, ok := strings.CutPrefix(format, "type."); ok {
		return local.makeTypFormatter(after, writer)
	}

	switch format {
	case "tags-path":
		output = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(
				writer,
				tl.GetObjectId(),
				&tl.Metadata.Cache.TagPaths,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "tags-path-with-types":
		output = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(
				writer,
				tl.GetObjectId(),
				&tl.Metadata.Cache.TagPaths,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "query-path":
		output = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(
				writer,
				tl.GetObjectId(),
				tl.Metadata.Cache.QueryPath,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "box":
		p := local.SkuFormatBoxTransactedNoColor()

		output = func(tl *sku.Transacted) (err error) {
			if _, err = p.EncodeStringTo(tl, writer); err != nil {
				err = errors.Wrap(err)
				return
			}

			if _, err = fmt.Fprintln(writer); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "box-archive":
		p := local.MakePrinterBoxArchive(
			writer,
			local.GetConfig().PrintOptions.PrintTime,
		)

		output = func(tl *sku.Transacted) (err error) {
			if err = p(tl); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "sha":
		output = func(tl *sku.Transacted) (err error) {
			_, err = fmt.Fprintln(writer, tl.Metadata.GetSha())
			return
		}

	case "sha-mutter":
		output = func(tl *sku.Transacted) (err error) {
			_, err = fmt.Fprintf(
				writer,
				"%s -> %s\n",
				tl.Metadata.GetSha(),
				tl.Metadata.GetMotherDigest(),
			)
			return
		}

	case "tags-all":
		output = func(tl *sku.Transacted) (err error) {
			for _, es := range tl.Metadata.Cache.TagPaths.Paths {
				if _, err = fmt.Fprintf(writer, "%s: %s\n", tl.GetObjectId(), es); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			for _, es := range tl.Metadata.Cache.TagPaths.All {
				if _, err = fmt.Fprintf(writer, "%s: %s -> %s\n", tl.GetObjectId(), es.Tag, es.Parents); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			return
		}

	case "tags-expanded":
		output = func(tl *sku.Transacted) (err error) {
			esImp := tl.GetMetadata().Cache.GetExpandedTags()
			// TODO-P3 determine if empty sets should be printed or not

			if _, err = fmt.Fprintln(
				writer,
				quiter.StringCommaSeparated(esImp),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "tags-implicit":
		output = func(tl *sku.Transacted) (err error) {
			esImp := tl.GetMetadata().Cache.GetImplicitTags()
			// TODO-P3 determine if empty sets should be printed or not

			if _, err = fmt.Fprintln(
				writer,
				quiter.StringCommaSeparated(esImp),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "tags":
		output = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(
				writer,
				quiter.StringCommaSeparated(
					tl.Metadata.GetTags(),
				),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "tags-newlines":
		output = func(tl *sku.Transacted) (err error) {
			if err = tl.Metadata.GetTags().EachPtr(func(e *ids.Tag) (err error) {
				_, err = fmt.Fprintln(writer, e)
				return
			}); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "description":
		output = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(writer, tl.GetMetadata().Description); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "text":
		formatter := typed_blob_store.MakeTextFormatter(
			local.GetStore().GetEnvRepo(),
			checkout_options.TextFormatterOptions{
				DoNotWriteEmptyDescription: true,
			},
			local.GetConfig(),
			checkout_mode.None,
		)

		output = func(tl *sku.Transacted) (err error) {
			_, err = formatter.EncodeStringTo(tl, writer)
			return
		}

	case "text-metadata_only":
		formatter := typed_blob_store.MakeTextFormatter(
			local.GetStore().GetEnvRepo(),
			checkout_options.TextFormatterOptions{
				DoNotWriteEmptyDescription: true,
			},
			local.GetConfig(),
			checkout_mode.MetadataOnly,
		)

		output = func(tl *sku.Transacted) (err error) {
			_, err = formatter.EncodeStringTo(tl, writer)
			return
		}

	case "object":
		fo := object_inventory_format.FormatForVersion(
			local.GetConfig().GetGenesisConfigPublic().GetStoreVersion(),
		)

		o := object_inventory_format.Options{
			Tai: true,
		}

		output = func(tl *sku.Transacted) (err error) {
			if _, err = fo.FormatPersistentMetadata(writer, tl, o); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "object-id-parent-tai":
		output = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintf(
				writer,
				"%s^@%s\n",
				&tl.ObjectId,
				tl.Metadata.Cache.ParentTai,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "object-id-sha":
		output = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintf(
				writer,
				"%s@%s\n",
				&tl.ObjectId,
				tl.GetObjectSha(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "object-id-blob-sha":
		output = func(tl *sku.Transacted) (err error) {
			ui.TodoP3("convert into an option")

			sh := tl.GetBlobSha()

			if sh.IsNull() {
				return
			}

			if _, err = fmt.Fprintf(
				writer,
				"%s %s\n",
				&tl.ObjectId,
				sh,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "object-id":
		output = func(e *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(
				writer,
				&e.ObjectId,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "object-id-abbreviated":
		output = func(e *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(
				writer,
				&e.ObjectId,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "object-id-tai":
		output = func(e *sku.Transacted) (err error) {
			_, err = fmt.Fprintln(writer, e.StringObjectIdTai())
			return
		}

	case "sku-metadata-sans-tai":
		output = func(e *sku.Transacted) (err error) {
			_, err = fmt.Fprintln(
				writer,
				sku_fmt.StringMetadataSansTai(e),
			)
			return
		}

	case "metadata":
		fo, err := object_inventory_format.FormatForKeyError(
			object_inventory_format.KeyFormatV5Metadata,
		)

		errors.PanicIfError(err)

		output = func(e *sku.Transacted) (err error) {
			_, err = fo.WriteMetadataTo(writer, e)
			return
		}

	case "metadata-plus-mutter":
		fo, err := object_inventory_format.FormatForKeyError(
			object_inventory_format.KeyFormatV5MetadataObjectIdParent,
		)

		errors.PanicIfError(err)

		output = func(e *sku.Transacted) (err error) {
			_, err = fo.WriteMetadataTo(writer, e)
			return
		}

	case "genre":
		output = func(e *sku.Transacted) (err error) {
			_, err = fmt.Fprintf(writer, "%s", e.GetObjectId().GetGenre())
			return
		}

	case "debug":
		output = func(e *sku.Transacted) (err error) {
			_, err = fmt.Fprintf(writer, "%#v\n", e)
			return
		}

	case "log":
		output = local.PrinterTransacted()

	case "json":
		enc := json.NewEncoder(writer)

		output = func(object *sku.Transacted) (err error) {
			var jsonRepresentation sku_fmt.Json

			if err = jsonRepresentation.FromTransacted(
				object,
				local.GetStore().GetEnvRepo().GetDefaultBlobStore(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = enc.Encode(jsonRepresentation); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "toml-json":
		enc := json.NewEncoder(writer)

		type tomlJson struct {
			sku_fmt.Json
			Blob map[string]any `json:"blob"`
		}

		output = func(object *sku.Transacted) (err error) {
			var jsonRep tomlJson

			if err = jsonRep.FromTransacted(
				object,
				local.GetStore().GetEnvRepo().GetDefaultBlobStore(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = toml.Unmarshal([]byte(jsonRep.Json.BlobString), &jsonRep.Blob); err != nil {
				err = nil

				if err = enc.Encode(jsonRep.Json); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
				// err = errors.Wrap(err)
				// return
			}

			if err = enc.Encode(jsonRep); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "json-toml-bookmark":
		enc := json.NewEncoder(writer)

		var resp client.ResponseWithParsedJSONBody

		req := client.BrowserRequest{
			Method: "GET",
			Path:   "/tabs",
		}

		var b client.BrowserProxy

		if err = b.Read(); err != nil {
			errors.PanicIfError(err)
		}

		if resp, err = b.Request(req); err != nil {
			errors.PanicIfError(err)
		}

		tabs := resp.ParsedJSONBody.([]interface{})

		output = func(o *sku.Transacted) (err error) {
			var j sku_fmt.JsonWithUrl

			if j, err = sku_fmt.MakeJsonTomlBookmark(
				o,
				local.GetStore().GetEnvRepo(),
				tabs,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = enc.Encode(j); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "tai":
		output = func(o *sku.Transacted) (err error) {
			fmt.Fprintln(writer, o.GetTai())
			return
		}

	case "blob":
		output = func(o *sku.Transacted) (err error) {
			var readCloser interfaces.ReadCloseDigester

			if readCloser, err = local.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
				o.GetBlobSha(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			defer errors.DeferredCloser(&err, readCloser)

			if _, err = io.Copy(writer, readCloser); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "text-sku-prefix":
		cliFmt := local.SkuFormatBoxTransactedNoColor()

		output = func(object *sku.Transacted) (err error) {
			sb := &strings.Builder{}

			if _, err = cliFmt.EncodeStringTo(object, sb); err != nil {
				err = errors.Wrap(err)
				return
			}

			if local.GetConfig().IsInlineType(object.GetType()) {
				var readCloser interfaces.ReadCloseDigester

				if readCloser, err = local.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
					object.GetBlobSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer errors.DeferredCloser(&err, readCloser)

				if _, err = delim_io.CopyWithPrefixOnDelim(
					'\n',
					sb.String(),
					local.GetOut(),
					readCloser,
					true,
				); err != nil {
					err = errors.Wrap(err)
					return
				}
			} else {
				if _, err = io.WriteString(writer, sb.String()); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			return
		}

	case "blob-sku-prefix":
		cliFmt := local.SkuFormatBoxTransactedNoColor()

		output = func(o *sku.Transacted) (err error) {
			var readCloser interfaces.ReadCloseDigester

			if readCloser, err = local.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
				o.GetBlobSha(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			defer errors.DeferredCloser(&err, readCloser)

			sb := &strings.Builder{}

			if _, err = cliFmt.EncodeStringTo(o, sb); err != nil {
				err = errors.Wrap(err)
				return
			}

			if _, err = delim_io.CopyWithPrefixOnDelim(
				'\n',
				sb.String(),
				local.GetOut(),
				readCloser,
				true,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "shas":
		output = func(z *sku.Transacted) (err error) {
			_, err = fmt.Fprintln(writer, &z.Metadata.Shas)
			return
		}

	case "mutter-sha":
		output = func(z *sku.Transacted) (err error) {
			_, err = fmt.Fprintln(writer, z.Metadata.GetMotherDigest())
			return
		}

	case "probe-shas":
		output = func(z *sku.Transacted) (err error) {
			sh1 := sha.FromStringContent(z.GetObjectId().String())
			sh2 := sha.FromStringContent(
				z.GetObjectId().String() + z.GetTai().String(),
			)
			defer digests.PutDigest(sh1)
			defer digests.PutDigest(sh2)
			_, err = fmt.Fprintln(writer, z.GetObjectId(), sh1, sh2)
			return
		}

	case "mutter":
		p := local.PrinterTransacted()

		output = func(z *sku.Transacted) (err error) {
			if z.Metadata.GetMotherDigest().IsNull() {
				return
			}

			if z, err = local.GetStore().GetStreamIndex().ReadOneObjectIdTai(
				z.GetObjectId(),
				z.Metadata.Cache.ParentTai,
			); err != nil {
				fmt.Fprintln(writer, err)
				err = nil
				return
			}

			return p(z)
		}

	case "inventory-list":
		p := local.MakePrinterBoxArchive(local.GetUIFile(), true)

		output = func(o *sku.Transacted) (err error) {
			if err = p(o); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "inventory-list-sans-tai":
		p := local.MakePrinterBoxArchive(local.GetUIFile(), false)

		output = func(o *sku.Transacted) (err error) {
			if err = p(o); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "blob-sha":
		output = func(o *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(writer, o.GetBlobSha()); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "type":
		output = func(o *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(writer, o.GetType().String()); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "verzeichnisse":
		p := local.PrinterTransacted()

		output = func(o *sku.Transacted) (err error) {
			sk := sku.GetTransactedPool().Get()
			defer sku.GetTransactedPool().Put(sk)

			if err = local.GetStore().GetStreamIndex().ReadOneObjectId(
				o.GetObjectId(),
				sk,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = p(sk); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "json-blob":
		e := json.NewEncoder(writer)

		output = func(o *sku.Transacted) (err error) {
			var a map[string]interface{}

			var readCloser interfaces.ReadCloseDigester

			if readCloser, err = local.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
				o.GetBlobSha(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			defer errors.Deferred(&err, readCloser.Close)

			d := toml.NewDecoder(readCloser)

			if err = d.Decode(&a); err != nil {
				ui.Err().Printf("%s: %s", o, err)
				err = nil
				return
			}

			a["description"] = o.Metadata.Description.String()
			a["identifier"] = o.ObjectId.String()

			if err = e.Encode(&a); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "toml":
		ui.TodoP3("limit to only zettels supporting toml")
		output = func(o *sku.Transacted) (err error) {
			var a map[string]interface{}

			var readCloser interfaces.ReadCloseDigester

			if readCloser, err = local.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
				o.GetBlobSha(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			defer errors.DeferredCloser(&err, readCloser)

			d := toml.NewDecoder(readCloser)

			if err = d.Decode(&a); err != nil {
				ui.Err().Printf("%s: %s", o, err)
				err = nil
				return
			}

			a["description"] = o.Metadata.Description.String()
			a["identifier"] = o.ObjectId.String()

			e := toml.NewEncoder(writer)

			if err = e.Encode(&a); err != nil {
				err = errors.Wrap(err)
				return
			}

			if _, err = writer.Write([]byte("\x00")); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "debug-sku-metadata":
		output = func(e *sku.Transacted) (err error) {
			_, err = fmt.Fprintln(
				writer,
				sku.StringMetadataTai(e),
			)
			return
		}

	case "debug-sku":
		output = func(e *sku.Transacted) (err error) {
			_, err = fmt.Fprintln(writer, sku.StringTaiGenreObjectIdShaBlob(e))
			return
		}

	default:
		err = MakeErrUnsupportedFormatterValue(format, genres.None)
	}

	return
}

func (local *Repo) makeTypFormatter(
	v string,
	out io.Writer,
) (f interfaces.FuncIter[*sku.Transacted], err error) {
	typeBlobStore := local.GetStore().GetTypedBlobStore().Type

	if out == nil {
		out = local.GetUIFile()
	}

	switch v {
	case "formatters":
		f = func(o *sku.Transacted) (err error) {
			var tt *sku.Transacted

			if tt, err = local.GetStore().ReadTransactedFromObjectId(o.GetType()); err != nil {
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

			if _, err = lw.WriteTo(out); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "formatter-uti-groups":
		fo := sku_fmt.MakeFormatterTypFormatterUTIGroups(
			local.GetStore(),
			typeBlobStore,
		)

		f = func(o *sku.Transacted) (err error) {
			if _, err = fo.Format(out, o); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	case "hooks.on_pre_commit":
		f = func(o *sku.Transacted) (err error) {
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

			if vp, err = local.GetStore().MakeLuaVMPoolV1(o, script); err != nil {
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

			local.GetUI().Print(f.String())

			return
		}

	case "vim-syntax-type":
		f = func(o *sku.Transacted) (err error) {
			var t *sku.Transacted

			if t, err = local.GetStore().ReadTransactedFromObjectId(o.GetType()); err != nil {
				if collections.IsErrNotFound(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
					return
				}
			}

			if t == nil || t.ObjectId.IsEmpty() || t.GetBlobSha().IsNull() {
				ty := ""

				switch o.GetGenre() {
				case genres.Type, genres.Tag, genres.Repo, genres.Config:
					ty = "toml"

				default:
					// TODO zettel default typ
				}

				if _, err = fmt.Fprintln(out, ty); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}

			var ta type_blobs.Blob

			if ta, _, err = typeBlobStore.ParseTypedBlob(
				t.GetType(),
				t.GetBlobSha(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			defer typeBlobStore.PutTypedBlob(t.GetType(), ta)

			if _, err = fmt.Fprintln(
				out,
				ta.GetVimSyntaxType(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

	default:
		err = MakeErrUnsupportedFormatterValue(
			v,
			genres.Type,
		)

		return
	}

	return
}
