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
		OverwriteSignatures bool

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
		&options.PrintCopies,
		"print-copies",
		true,
		"output when blobs are copied",
	)

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
		"which object genres should have their blobs copied",
	)

	flagDefinitions.BoolVar(
		&options.OverwriteSignatures,
		"overwrite-signatures",
		false,
		"ignore object pubkeys and signatures and generate new ones (causing this repo to create the objects as new instead of importing them)",
	)
}

func (options ImporterOptions) WithPrintCopies(
	value bool,
) ImporterOptions {
	options.PrintCopies = value
	return options
}
