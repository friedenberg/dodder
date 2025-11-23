package object_metadata

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
)

type Dependencies struct {
	EnvDir        env_dir.Env
	BlobStore     interfaces.BlobStore
	BlobFormatter script_config.RemoteScript
}

func (deps Dependencies) GetBlobDigestType() interfaces.FormatHash {
	hashType := deps.BlobStore.GetDefaultHashType()

	if hashType == nil {
		panic("no hash type set")
	}

	return hashType
}

func (deps Dependencies) writeComments(
	writer io.Writer,
	context TextFormatterContext,
) (n int64, err error) {
	n1 := 0

	for comment := range context.GetMetadataMutable().GetComments() {
		n1, err = io.WriteString(writer, "% ")
		n += int64(n1)

		if err != nil {
			return n, err
		}

		n1, err = io.WriteString(writer, comment)
		n += int64(n1)

		if err != nil {
			return n, err
		}

		n1, err = io.WriteString(writer, "\n")
		n += int64(n1)

		if err != nil {
			return n, err
		}
	}

	return n, err
}

func (deps Dependencies) writeBoundary(
	writer io.Writer,
	_ TextFormatterContext,
) (n int64, err error) {
	return ohio.WriteLine(writer, triple_hyphen_io.Boundary)
}

func (deps Dependencies) writeNewLine(
	writer io.Writer,
	_ TextFormatterContext,
) (n int64, err error) {
	return ohio.WriteLine(writer, "")
}

func (deps Dependencies) writeCommonMetadataFormat(
	writer io.Writer,
	formatterContext TextFormatterContext,
) (n int64, err error) {
	lineWriter := format.NewLineWriter()

	metadata := formatterContext.GetMetadata()
	description := metadata.GetDescription()

	if description.String() != "" || !formatterContext.DoNotWriteEmptyDescription {
		reader, repool := pool.GetStringReader(description.String())
		defer repool()

		stringReader := bufio.NewReader(reader)

		for {
			var line string
			line, err = stringReader.ReadString('\n')
			isEOF := err == io.EOF

			if err != nil && !isEOF {
				err = errors.Wrap(err)
				return n, err
			}

			lineWriter.WriteLines(
				fmt.Sprintf("# %s", strings.TrimSpace(line)),
			)

			if isEOF {
				break
			}
		}
	}

	for _, e := range quiter.SortedValues(metadata.GetTags()) {
		if ids.IsEmpty(e) {
			continue
		}

		lineWriter.WriteFormat("- %s", e)
	}

	if n, err = lineWriter.WriteTo(writer); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (deps Dependencies) writeTypeAndSigIfNecessary(
	writer io.Writer,
	formatterContext TextFormatterContext,
) (n int64, err error) {
	metadata := formatterContext.GetMetadata()
	tipe := metadata.GetType()
	typeSig := metadata.GetLockfile().GetType()

	if tipe.IsEmpty() {
		return n, err
	}

	if typeSig.IsEmpty() {
		return ohio.WriteLine(
			writer,
			fmt.Sprintf(
				"! %s",
				tipe.StringSansOp(),
			),
		)
	}

	return deps.writeTypeAndSig(writer, formatterContext)
}

func (deps Dependencies) writeTypeAndSig(
	writer io.Writer,
	formatterContext TextFormatterContext,
) (n int64, err error) {
	metadata := formatterContext.GetMetadata()
	tipe := metadata.GetType()
	typeSig := metadata.GetLockfile().GetType()

	if tipe.IsEmpty() {
		return n, err
	}

	if typeSig.IsEmpty() {
		err = errors.Errorf("empty type signature for type: %q", tipe)
		return n, err
	}

	return ohio.WriteLine(
		writer,
		fmt.Sprintf(
			"! %s@%s",
			tipe.StringSansOp(),
			typeSig,
		),
	)
}

func (deps Dependencies) writeBlobDigest(
	writer io.Writer,
	formatterContext TextFormatterContext,
) (n int64, err error) {
	metadata := formatterContext.GetMetadata()

	return ohio.WriteLine(
		writer,
		fmt.Sprintf(
			"@ %s",
			metadata.GetBlobDigest(),
		),
	)
}

func (deps Dependencies) writeBlobPath(
	writer io.Writer,
	formatterContext TextFormatterContext,
) (n int64, err error) {
	var blobPath string

	metadata := formatterContext.PersistentFormatterContext.GetMetadataMutable()

	for field := range metadata.GetIndexMutable().GetFields() {
		if strings.ToLower(field.Key) == "blob" {
			blobPath = field.Value
			break
		}
	}

	if blobPath != "" {
		blobPath = deps.EnvDir.RelToCwdOrSame(blobPath)
	} else {
		err = errors.ErrorWithStackf("path not found in fields")
		return n, err
	}

	return ohio.WriteLine(writer, fmt.Sprintf("@ %s", blobPath))
}

func (deps Dependencies) writeBlob(
	writer io.Writer,
	formatterContext TextFormatterContext,
) (n int64, err error) {
	var blobReader interfaces.BlobReader

	metadata := formatterContext.GetMetadata()

	if blobReader, err = deps.BlobStore.MakeBlobReader(
		metadata.GetBlobDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	if blobReader == nil {
		err = errors.ErrorWithStackf("blob reader is nil")
		return n, err
	}

	defer errors.DeferredCloser(&err, blobReader)

	if deps.BlobFormatter != nil {
		var writerTo io.WriterTo

		if writerTo, err = script_config.MakeWriterToWithStdin(
			deps.BlobFormatter,
			deps.EnvDir.MakeCommonEnv(),
			blobReader,
		); err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		if n, err = writerTo.WriteTo(writer); err != nil {
			var errExit *exec.ExitError

			if errors.As(err, &errExit) {
				err = MakeErrBlobFormatterFailed(errExit)
			}

			err = errors.Wrap(err)

			return n, err
		}
	} else {
		if n, err = io.Copy(writer, blobReader); err != nil {
			err = errors.Wrap(err)
			return n, err
		}
	}

	return n, err
}
