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
		DedupingFormatId    string
		BlobGenres          ids.Genre
		ExcludeObjects      bool
		RemoteBlobStore     interfaces.BlobStore
		PrintCopies         bool
		AllowMergeConflicts bool
		BlobCopierDelegate  interfaces.FuncIter[sku.BlobCopyResult]
		ParentNegotiator    sku.ParentNegotiator
		CheckedOutPrinter   interfaces.FuncIter[*sku.CheckedOut]
	}
)

// TODO add HTTP header options for these flags
type RemoteTransferOptions struct {
	PrintCopies         bool
	BlobGenres          ids.Genre
	IncludeObjects      bool
	IncludeBlobs        bool
	AllowMergeConflicts bool
}

func (options *RemoteTransferOptions) SetFlagSet(
	flagDefinitions interfaces.CommandLineFlagDefinitions,
) {
	flagDefinitions.BoolVar(
		&options.IncludeObjects,
		"include-objects",
		true,
		"imports the object during transfer",
	)

	flagDefinitions.BoolVar(
		&options.IncludeBlobs,
		"include-blobs",
		true,
		"copy the blob when performing the object transfer",
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

func (options RemoteTransferOptions) WithPrintCopies(
	value bool,
) RemoteTransferOptions {
	options.PrintCopies = value
	return options
}
