package repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	Importer interface {
		GetCheckedOutPrinter() interfaces.FuncIter[*sku.CheckedOut]

		SetCheckedOutPrinter(
			p interfaces.FuncIter[*sku.CheckedOut],
		)

		ImportBlobIfNecessary(
			sk *sku.Transacted,
		) (err error)

		Import(
			external *sku.Transacted,
		) (co *sku.CheckedOut, err error)
	}

	ImporterOptions struct {
		BlobGenres          ids.Genre
		PrintCopies         bool
		ExcludeObjects      bool
		ExcludeBlobs        bool
		AllowMergeConflicts bool

		DedupingFormatId   string
		RemoteBlobStore    interfaces.BlobStore
		BlobCopierDelegate interfaces.FuncIter[sku.BlobCopyResult]
		ParentNegotiator   sku.ParentNegotiator
		CheckedOutPrinter  interfaces.FuncIter[*sku.CheckedOut]
	}
)

// TODO add HTTP header options for these flags
func (options *ImporterOptions) SetFlagSet(
	flagDefinitions interfaces.CommandLineFlagDefinitions,
) {
	flagDefinitions.BoolVar(
		&options.ExcludeObjects,
		"exclude-objects",
		false,
		"excludes objects during transfer",
	)

	flagDefinitions.BoolVar(
		&options.ExcludeBlobs,
		"exclude-blobs",
		false,
		"excludes blobs during the remote transfer",
	)

	flagDefinitions.BoolVar(
		&options.AllowMergeConflicts,
		"allow-merge-conflicts",
		false,
		"ignore merge conflicts and allow incompatible histories to coexist",
	)

	flagDefinitions.Var(
		&options.BlobGenres,
		"blob-genres",
		"which blob genres should have their blobs copied",
	)
}

func (options ImporterOptions) WithPrintCopies(
	value bool,
) ImporterOptions {
	options.PrintCopies = value
	return options
}
