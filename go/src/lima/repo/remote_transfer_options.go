package repo

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// TODO add HTTP header options for these flags
type RemoteTransferOptions struct {
	PrintCopies         bool
	BlobGenres          ids.Genre
	IncludeObjects      bool
	IncludeBlobs        bool
	AllowMergeConflicts bool
}

func (options *RemoteTransferOptions) SetFlagSet(f *flag.FlagSet) {
	f.BoolVar(
		&options.IncludeObjects,
		"include-objects",
		true,
		"imports the object during transfer",
	)

	f.BoolVar(
		&options.IncludeBlobs,
		"include-blobs",
		true,
		"copy the blob when performing the object transfer",
	)

	f.BoolVar(
		&options.AllowMergeConflicts,
		"allow-merge-conflicts",
		false,
		"ignore merge conflicts and allow incompatible histories to coexist",
	)

	f.Var(
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
