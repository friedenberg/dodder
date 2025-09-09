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
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_io"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_fmt"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_json_fmt"
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
				slices.Sorted(maps.Keys(formatters)),
			)
		} else {
			return fmt.Sprintf(
				"%q",
				slices.Sorted(maps.Keys(formatters)),
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
		err = flags.ErrInvalidValue{
			Actual:   value,
			Expected: slices.Sorted(maps.Keys(formatters)),
		}

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
				repo.GetConfig().GetPrintOptions().BoxPrintTime,
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
	"merkle-repo-pubkey": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.Metadata.GetRepoPubKey().StringWithFormat(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"merkle-object": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.Metadata.GetObjectDigest().StringWithFormat(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"merkle-mother": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintf(
					writer,
					"%q -> %q\n",
					object.Metadata.GetObjectDigest().StringWithFormat(),
					object.Metadata.GetMotherObjectSig().StringWithFormat(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

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
	"object-id-digest": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintf(
					writer,
					"%s@%s\n",
					&object.ObjectId,
					object.GetObjectDigest().StringWithFormat(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"object-id-blob-digest": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				ui.TodoP3("convert into an option")

				digest := object.GetBlobDigest()

				if digest.IsNull() {
					return
				}

				if _, err = fmt.Fprintf(
					writer,
					"%s %s\n",
					&object.ObjectId,
					digest.StringWithFormat(),
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
				var jsonRep sku_json_fmt.Transacted

				if err = jsonRep.FromTransacted(
					object,
					repo.GetStore().GetEnvRepo().GetDefaultBlobStore(),
				); err != nil {
					err = errors.Wrap(err)
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
	"toml-json": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			enc := json.NewEncoder(writer)

			type tomlJson struct {
				sku_json_fmt.Transacted
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

				if err = toml.Unmarshal([]byte(
					jsonRep.Transacted.BlobString),
					&jsonRep.Blob,
				); err != nil {
					err = nil

					if err = enc.Encode(jsonRep.Transacted); err != nil {
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
				var objectJSON sku_json_fmt.JsonWithUrl

				if objectJSON, err = sku_json_fmt.MakeJsonTomlBookmark(
					object,
					repo.GetStore().GetEnvRepo(),
					tabs,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = enc.Encode(objectJSON); err != nil {
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
				var readCloser interfaces.ReadCloseMarklIdGetter

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().MakeBlobReader(
					object.GetBlobDigest(),
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
	"text-box-prefix": {
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
					var readCloser interfaces.ReadCloseMarklIdGetter

					if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().MakeBlobReader(
						object.GetBlobDigest(),
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
	"blob-box-prefix": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			cliFmt := repo.SkuFormatBoxTransactedNoColor()

			return func(object *sku.Transacted) (err error) {
				var readCloser interfaces.ReadCloseMarklIdGetter

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().MakeBlobReader(
					object.GetBlobDigest(),
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
	"merkle": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			// TODO
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintln(writer, &object.Metadata)
				return
			}
		},
	},
	"sig-bytes-hex": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintf(
					writer,
					"%x\n",
					object.Metadata.GetObjectSig().GetBytes(),
				)
				return
			}
		},
	},
	"sig-mother-bytes-hex": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				_, err = fmt.Fprintf(
					writer,
					"%x\n",
					object.Metadata.GetMotherObjectSig().GetBytes(),
				)
				return
			}
		},
	},
	"sig-mother": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.Metadata.GetMotherObjectSig(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"merkle-probe": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				for probeId := range object.AllProbeIds() {
					if _, err = fmt.Fprintf(
						writer,
						"%s (%q): %q -> %q\n",
						object.GetObjectId(),
						probeId.Key,
						"TODO value",
						probeId.Id.StringWithFormat(),
					); err != nil {
						err = errors.Wrap(err)
						return
					}
				}

				return
			}
		},
	},
	"mother": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.PrinterTransacted()

			return func(object *sku.Transacted) (err error) {
				if object.Metadata.GetMotherObjectSig().IsNull() {
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
	"inventory_list": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			funcPrint := repo.MakePrinterBoxArchive(repo.GetUIFile(), true)

			return func(object *sku.Transacted) (err error) {
				if err = funcPrint(object); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"inventory_list-sans-tai": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			funcPrint := repo.MakePrinterBoxArchive(repo.GetUIFile(), false)

			return func(object *sku.Transacted) (err error) {
				if err = funcPrint(object); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		},
	},
	"merkle-sig": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.Metadata.GetObjectSig().StringWithFormat(),
				); err != nil {
					err = errors.Wrap(err)
					return
				}
				return
			}
		},
	},
	"merkle-blob": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			return func(object *sku.Transacted) (err error) {
				if _, err = fmt.Fprintln(
					writer,
					object.GetBlobDigest().StringWithFormat(),
				); err != nil {
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
	"index": {
		FormatFuncConstructor: func(
			repo *Repo,
			writer interfaces.WriterAndStringWriter,
		) interfaces.FuncIter[*sku.Transacted] {
			p := repo.PrinterTransacted()

			return func(object *sku.Transacted) (err error) {
				indexObject := sku.GetTransactedPool().Get()
				defer sku.GetTransactedPool().Put(indexObject)

				if err = repo.GetStore().GetStreamIndex().ReadOneObjectId(
					object.GetObjectId(),
					indexObject,
				); err != nil {
					err = errors.Wrap(err)
					return
				}

				if err = p(indexObject); err != nil {
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
				var a map[string]any

				var readCloser interfaces.ReadCloseMarklIdGetter

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().MakeBlobReader(
					object.GetBlobDigest(),
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
				var a map[string]any

				var readCloser interfaces.ReadCloseMarklIdGetter

				if readCloser, err = repo.GetStore().GetEnvRepo().GetDefaultBlobStore().MakeBlobReader(
					object.GetBlobDigest(),
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
					sku.StringMetadataTaiMerkle(object),
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
					sku.StringTaiGenreObjectIdObjectDigestBlobDigest(object),
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
