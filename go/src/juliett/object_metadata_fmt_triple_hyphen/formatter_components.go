package object_metadata_fmt_triple_hyphen

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
)

type formatterComponents Factory

func (factory formatterComponents) getWriteTypeAndSigFunc() funcWrite {
	if factory.AllowMissingTypeSig || true {
		return factory.writeTypeAndSigIfNecessary
	} else {
		return factory.writeTypeAndSig
	}
}

func (factory formatterComponents) writeComments(
	writer interfaces.WriterAndStringWriter,
	context FormatterContext,
) (n int64, err error) {
	n1 := 0

	for comment := range context.GetMetadata().GetIndex().GetComments() {
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

func (factory formatterComponents) writeBoundary(
	writer interfaces.WriterAndStringWriter,
	_ FormatterContext,
) (n int64, err error) {
	return ohio.WriteLine(writer, triple_hyphen_io.Boundary)
}

func (factory formatterComponents) writeNewLine(
	writer interfaces.WriterAndStringWriter,
	_ FormatterContext,
) (n int64, err error) {
	return ohio.WriteLine(writer, "")
}

func (factory formatterComponents) writeCommonMetadataFormat(
	writer interfaces.WriterAndStringWriter,
	formatterContext FormatterContext,
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

	for _, tag := range quiter.SortedValues(metadata.GetTags().All()) {
		if ids.IsEmpty(tag) {
			continue
		}

		lineWriter.WriteFormat("- %s", tag)
	}

	if n, err = lineWriter.WriteTo(writer); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (factory formatterComponents) writeTypeAndSigIfNecessary(
	writer interfaces.WriterAndStringWriter,
	formatterContext FormatterContext,
) (n int64, err error) {
	metadata := formatterContext.GetMetadata()
	typeTuple := metadata.GetTypeLock()

	if typeTuple.Key.IsEmpty() {
		return n, err
	}

	if typeTuple.Value.IsEmpty() {
		return ohio.WriteLine(
			writer,
			fmt.Sprintf(
				"! %s",
				typeTuple.Key.StringSansOp(),
			),
		)
	}

	return factory.writeTypeAndSig(writer, formatterContext)
}

func (factory formatterComponents) writeTypeAndSig(
	writer interfaces.WriterAndStringWriter,
	formatterContext FormatterContext,
) (n int64, err error) {
	metadata := formatterContext.GetMetadata()
	typeTuple := metadata.GetTypeLock()

	if typeTuple.Key.IsEmpty() {
		return n, err
	}

	if typeTuple.Value.IsEmpty() {
		err = errors.Errorf("empty type signature for type: %q", typeTuple.Key)
		return n, err
	}

	return ohio.WriteLine(
		writer,
		fmt.Sprintf(
			"! %s@%s",
			typeTuple.Key.StringSansOp(),
			typeTuple.Value,
		),
	)
}

func (factory formatterComponents) writeBlobDigest(
	writer interfaces.WriterAndStringWriter,
	formatterContext FormatterContext,
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

func (factory formatterComponents) writeBlobPath(
	writer interfaces.WriterAndStringWriter,
	formatterContext FormatterContext,
) (n int64, err error) {
	var blobPath string

	metadata := formatterContext.GetMetadataMutable()

	for field := range metadata.GetIndexMutable().GetFields() {
		if strings.ToLower(field.Key) == "blob" {
			blobPath = field.Value
			break
		}
	}

	if blobPath != "" {
		blobPath = factory.EnvDir.RelToCwdOrSame(blobPath)
	} else {
		err = errors.ErrorWithStackf("path not found in fields")
		return n, err
	}

	return ohio.WriteLine(writer, fmt.Sprintf("@ %s", blobPath))
}

func (factory formatterComponents) writeBlob(
	writer interfaces.WriterAndStringWriter,
	formatterContext FormatterContext,
) (n int64, err error) {
	var blobReader interfaces.BlobReader

	metadata := formatterContext.GetMetadata()

	if blobReader, err = factory.BlobStore.MakeBlobReader(
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

	if factory.BlobFormatter != nil {
		var writerTo io.WriterTo

		if writerTo, err = script_config.MakeWriterToWithStdin(
			factory.BlobFormatter,
			factory.EnvDir.MakeCommonEnv(),
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
