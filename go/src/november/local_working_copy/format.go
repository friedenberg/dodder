package local_working_copy

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
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
	"code.linenisgreat.com/dodder/go/src/charlie/delim_io"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_fmt"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

type (
	FormatFuncConstructor func(
		*Repo,
		// TODO consider switching to using fd.Std
		interfaces.WriterAndStringWriter,
	) interfaces.FuncIter[*sku.Transacted]

	FormatFuncConstructorEntry struct {
		Name        string
		description string
		FormatFuncConstructor
	}

	FormatFlag struct {
		DefaultFormatter FormatFuncConstructorEntry

		wasSet    bool
		formatter FormatFuncConstructorEntry
	}
)

func (formatFlag *FormatFlag) WasSet() bool {
	return formatFlag.wasSet
}

func (formatFlag *FormatFlag) GetName() string {
	return formatFlag.formatter.Name
}

func (formatFlag *FormatFlag) String() string {
	if formatFlag.formatter.Name == "" {
		if formatFlag.DefaultFormatter.Name != "" {
			return fmt.Sprintf(
				"Default: %s, All: %q",
				formatFlag.DefaultFormatter.Name,
				slices.Collect(maps.Keys(formatters)),
			)
		} else {
			return fmt.Sprintf(
				"%q",
				slices.Collect(maps.Keys(formatters)),
			)
		}
	} else if formatFlag.formatter.description != "" {
		return fmt.Sprintf("%s: %s", formatFlag.formatter.Name, formatFlag.formatter.description)
	} else {
		return formatFlag.formatter.Name
	}
}

var formatterCompletions = func() map[string]string {
	completion := make(map[string]string, len(formatters))

	for name, entry := range formatters {
		if entry.description != "" {
			completion[name] = name
		} else {
			completion[name] = entry.description
		}
	}

	return completion
}()

func (formatFlag *FormatFlag) GetCLICompletion() map[string]string {
	return formatterCompletions
}

func (formatFlag *FormatFlag) Set(value string) (err error) {
	var ok bool
	var entry FormatFuncConstructorEntry

	if entry, ok = formatters[value]; !ok {
		err = errors.BadRequestf(
			"unsupported format. Available formats: %q",
			value,
			slices.Collect(maps.Keys(formatters)),
		)

		return
	}

	formatFlag.wasSet = true
	entry.Name = value
	formatFlag.formatter = entry

	return
}

func (formatFlag *FormatFlag) MakeFormatFunc(
	repo *Repo,
	writer interfaces.WriterAndStringWriter,
) interfaces.FuncIter[*sku.Transacted] {
	if formatFlag.formatter.Name == "" &&
		formatFlag.DefaultFormatter.Name == "" {
		errors.ContextCancelWithErrorf(
			repo,
			"neither format flag nor default were set",
		)
		return nil
	} else if formatFlag.formatter.Name == "" {
		return formatFlag.DefaultFormatter.FormatFuncConstructor(repo, writer)
	}

	return formatFlag.formatter.FormatFuncConstructor(repo, writer)
}

func GetFormatFuncConstructorEntry(name string) FormatFuncConstructorEntry {
	entry, ok := formatters[name]

	if !ok {
		panic(
			fmt.Sprintf(
				"format name not found: %q. Available: %s",
				name,
				slices.Collect(maps.Keys(formatters)),
			),
		)
	}

	entry.Name = name

	return entry
}

var formatters = map[string]FormatFuncConstructorEntry{
	"tags-path": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.GetObjectId(),
					&object.Metadata.Cache.TagPaths,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"tags-path-with-types": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.GetObjectId(),
					&object.Metadata.Cache.TagPaths,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"query-path": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.GetObjectId(),
					object.Metadata.Cache.QueryPath,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"box": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.SkuFormatBoxTransactedNoColor()

			return func(object *sku.Transacted) (err error) {
				if _, err = p.EncodeStringTo(object, writer); err != nil {
					err = errors.Wrap(err)
					return
				}

				if _, err = fmt.Fprintln(writer); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"box-archive": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.MakePrinterBoxArchive(
				writer,
				repo.GetConfig().PrintOptions.PrintTime,
			)

			return func(object *sku.Transacted) (err error) {
				if err = p(object); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"sha": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(writer, object.Metadata.GetSha())
				return
			}
		},
	},
	"sha-mutter": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintf(
					writer,
					"%s -> %s\n",
					object.Metadata.GetSha(),
					object.Metadata.GetMotherDigest(),
				)
				return
			}
		},
	},
	"tags-all": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				for _, es := range object.Metadata.Cache.TagPaths.Paths {
					if _, err = fmt.Fprintf(writer, "%s: %s\n", object.GetObjectId(), es); err != nil {
						err = errors.Wrap(err)
						return
					}
				}

				for _, es := range object.Metadata.Cache.TagPaths.All {
					if _, err = fmt.Fprintf(writer, "%s: %s -> %s\n", object.GetObjectId(), es.Tag, es.Parents); err != nil {
						err = errors.Wrap(err)
						return
					}
				}

				return
			}
		},
	},
	"tags-expanded": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				esImp := object.GetMetadata().Cache.GetExpandedTags()

				if _, err = fmt.Fprintln(
					writer,
					quiter.StringCommaSeparated(esImp),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"tags-implicit": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				esImp := object.GetMetadata().Cache.GetImplicitTags()

				if _, err = fmt.Fprintln(
					writer,
					quiter.StringCommaSeparated(esImp),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"tags": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					quiter.StringCommaSeparated(
						object.Metadata.GetTags(),
					),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"tags-newlines": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				for e := range object.Metadata.GetTags().AllPtr() {
					if _, err = fmt.Fprintln(writer, e); err != nil {
						err = errors.Wrap(err)
						return
					}
				}

				return
			}
		},
	},
	"description": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(writer, object.GetMetadata().Description); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"text": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			formatter := typed_blob_store.MakeTextFormatter(
				repo.GetStore().GetEnvRepo(),
				checkout_options.TextFormatterOptions{
					DoNotWriteEmptyDescription: true,
				},
				repo.GetConfig(),
				checkout_mode.None,
			)

			return func(object *sku.Transacted) (err error) {
				_, err = formatter.EncodeStringTo(object, writer)
				return
			}
		},
	},
	"text-metadata_only": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			formatter := typed_blob_store.MakeTextFormatter(
				repo.GetStore().GetEnvRepo(),
				checkout_options.TextFormatterOptions{
					DoNotWriteEmptyDescription: true,
				},
				repo.GetConfig(),
				checkout_mode.MetadataOnly,
			)

			return func(object *sku.Transacted) (err error) {
				_, err = formatter.EncodeStringTo(object, writer)
				return
			}
		},
	},
	"object": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			fo := object_inventory_format.FormatForVersion(
				repo.GetConfig().GetGenesisConfigPublic().GetStoreVersion(),
			)

			o := object_inventory_format.Options{
				Tai: true,
			}

			return func(object *sku.Transacted) (err error) {
				if _, err = fo.FormatPersistentMetadata(writer, object, o); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"object-id-parent-tai": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintf(
					writer,
					"%s^@%s\n",
					&object.ObjectId,
					object.Metadata.Cache.ParentTai,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"object-id-sha": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintf(
					writer,
					"%s@%s\n",
					&object.ObjectId,
					object.GetObjectSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"object-id-blob-sha": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				ui.TodoP3("convert into an option")

				sh := object.GetBlobSha()

				if sh.IsNull() {
					return
				}

				if _, err = fmt.Fprintf(
					writer,
					"%s %s\n",
					&object.ObjectId,
					sh,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"object-id": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					&object.ObjectId,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"object-id-abbreviated": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					&object.ObjectId,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"object-id-tai": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(writer, object.StringObjectIdTai())
				return
			}
		},
	},
	"sku-metadata-sans-tai": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(
					writer,
					sku_fmt.StringMetadataSansTai(object),
				)
				return
			}
		},
	},
	"metadata": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			fo, err := object_inventory_format.FormatForKeyError(
				object_inventory_format.KeyFormatV5Metadata,
			)
			if err != nil {
				repo.Cancel(err)
				return nil
			}

			return func(object *sku.Transacted) (err error) {
				_, err = fo.WriteMetadataTo(writer, object)
				return
			}
		},
	},
	"metadata-plus-mutter": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			fo, err := object_inventory_format.FormatForKeyError(
				object_inventory_format.KeyFormatV5MetadataObjectIdParent,
			)
			if err != nil {
				repo.Cancel(err)
				return nil
			}

			return func(object *sku.Transacted) (err error) {
				_, err = fo.WriteMetadataTo(writer, object)
				return
			}
		},
	},
	"genre": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintf(
					writer,
					"%s",
					object.GetObjectId().GetGenre(),
				)
				return
			}
		},
	},
	"debug": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintf(writer, "%#v\n", object)
				return
			}
		},
	},
	"log": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return repo.PrinterTransacted()
		},
	},
	"json": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			enc := json.NewEncoder(writer)

			return func(object *sku.Transacted) (err error) {
				var jsonRepresentation sku_fmt.Json

				if err = jsonRepresentation.FromTransacted(
					object,
					repo.GetStore().GetEnvRepo().GetDefaultBlobStore(),
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
		},
	},
	"toml-json": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			enc := json.NewEncoder(writer)

			type tomlJson struct {
				sku_fmt.Json
				Blob map[string]any `json:"blob"`
			}

			return func(object *sku.Transacted) (err error) {
				var jsonRep tomlJson

				if err = jsonRep.FromTransacted(
					object,
					repo.GetStore().GetEnvRepo().GetDefaultBlobStore(),
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
				}

				if err = enc.Encode(jsonRep); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"json-toml-bookmark": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			enc := json.NewEncoder(writer)

			var resp client.ResponseWithParsedJSONBody
			var b client.BrowserProxy

			req := client.BrowserRequest{
				Method: "GET",
				Path:   "/tabs",
			}

			if err := b.Read(); err != nil {
				repo.Cancel(err)
				return nil
			}

			var err error
			if resp, err = b.Request(req); err != nil {
				repo.Cancel(err)
				return nil
			}

			tabs := resp.ParsedJSONBody.([]interface{})

			return func(object *sku.Transacted) (err error) {
				var j sku_fmt.JsonWithUrl

				if j, err = sku_fmt.MakeJsonTomlBookmark(
					object,
					repo.GetStore().GetEnvRepo(),
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
		},
	},
	"tai": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				fmt.Fprintln(writer, object.GetTai())
				return
			}
		},
	},
	"blob": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				var readCloser interfaces.ReadCloseDigester

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
					object.GetBlobSha(),
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
		},
	},
	"text-sku-prefix": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			cliFmt := repo.SkuFormatBoxTransactedNoColor()

			return func(object *sku.Transacted) (err error) {
				sb := &strings.Builder{}

				if _, err = cliFmt.EncodeStringTo(object, sb); err != nil {
					err = errors.Wrap(err)
					return
				}

				if repo.GetConfig().IsInlineType(object.GetType()) {
					var readCloser interfaces.ReadCloseDigester

					if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
						object.GetBlobSha(),
					); err != nil {
						err = errors.Wrap(err)
						return
					}

					defer errors.DeferredCloser(&err, readCloser)

					if _, err = delim_io.CopyWithPrefixOnDelim(
						'\n',
						sb.String(),
						repo.GetOut(),
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
		},
	},
	"blob-sku-prefix": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			cliFmt := repo.SkuFormatBoxTransactedNoColor()

			return func(object *sku.Transacted) (err error) {
				var readCloser interfaces.ReadCloseDigester

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
					object.GetBlobSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer errors.DeferredCloser(&err, readCloser)

				sb := &strings.Builder{}

				if _, err = cliFmt.EncodeStringTo(object, sb); err != nil {
					err = errors.Wrap(err)
					return
				}

				if _, err = delim_io.CopyWithPrefixOnDelim(
					'\n',
					sb.String(),
					repo.GetOut(),
					readCloser,
					true,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"shas": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(writer, &object.Metadata.Shas)
				return
			}
		},
	},
	"mutter-sha": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(writer, object.Metadata.GetMotherDigest())
				return
			}
		},
	},
	"probe-shas": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				sh1 := sha.FromStringContent(object.GetObjectId().String())
				sh2 := sha.FromStringContent(
					object.GetObjectId().String() + object.GetTai().String(),
				)
				defer digests.PutBlobId(sh1)
				defer digests.PutBlobId(sh2)
				_, err = fmt.Fprintln(writer, object.GetObjectId(), sh1, sh2)
				return
			}
		},
	},
	"mutter": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.PrinterTransacted()

			return func(object *sku.Transacted) (err error) {
				if object.Metadata.GetMotherDigest().IsNull() {
					return
				}

				if object, err = repo.GetStore().GetStreamIndex().ReadOneObjectIdTai(
					object.GetObjectId(),
					object.Metadata.Cache.ParentTai,
				); err != nil {
					fmt.Fprintln(writer, err)
					err = nil
					return
				}

				return p(object)
			}
		},
	},
	"inventory-list": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.MakePrinterBoxArchive(repo.GetUIFile(), true)

			return func(object *sku.Transacted) (err error) {
				if err = p(object); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"inventory-list-sans-tai": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.MakePrinterBoxArchive(repo.GetUIFile(), false)

			return func(object *sku.Transacted) (err error) {
				if err = p(object); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"blob-sha": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(writer, object.GetBlobSha()); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"type": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(writer, object.GetType().String()); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"verzeichnisse": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.PrinterTransacted()

			return func(object *sku.Transacted) (err error) {
				sk := sku.GetTransactedPool().Get()
				defer sku.GetTransactedPool().Put(sk)

				if err = repo.GetStore().GetStreamIndex().ReadOneObjectId(
					object.GetObjectId(),
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
		},
	},
	"json-blob": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			e := json.NewEncoder(writer)

			return func(object *sku.Transacted) (err error) {
				var a map[string]interface{}

				var readCloser interfaces.ReadCloseDigester

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
					object.GetBlobSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer errors.Deferred(&err, readCloser.Close)

				d := toml.NewDecoder(readCloser)

				if err = d.Decode(&a); err != nil {
					ui.Err().Printf("%s: %s", object, err)
					err = nil
					return
				}

				a["description"] = object.Metadata.Description.String()
				a["identifier"] = object.ObjectId.String()

				if err = e.Encode(&a); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"toml": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				ui.TodoP3("limit to only zettels supporting toml")
				var a map[string]interface{}

				var readCloser interfaces.ReadCloseDigester

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().BlobReader(
					object.GetBlobSha(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				defer errors.DeferredCloser(&err, readCloser)

				d := toml.NewDecoder(readCloser)

				if err = d.Decode(&a); err != nil {
					ui.Err().Printf("%s: %s", object, err)
					err = nil
					return
				}

				a["description"] = object.Metadata.Description.String()
				a["identifier"] = object.ObjectId.String()

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
		},
	},
	"debug-sku-metadata": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(
					writer,
					sku.StringMetadataTai(object),
				)
				return
			}
		},
	},
	"debug-sku": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(
					writer,
					sku.StringTaiGenreObjectIdShaBlob(object),
				)
				return
			}
		},
	},
}

func init() {
	for name, entry := range typeFormatters {
		formatters[fmt.Sprintf("type.%s", name)] = makeFormatEntryFromTypeFormat(
			entry,
		)
	}
}
