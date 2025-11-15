package object_metadata

import (
	"io"
	"path"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_reader"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
)

type textParser2 struct {
	interfaces.BlobWriterFactory
	TextParserContext
	hashType interfaces.FormatHash
	Blob     fd.FD
}

func (parser *textParser2) ReadFrom(r io.Reader) (n int64, err error) {
	metadata := parser.GetMetadataMutable()
	Resetter.Reset(metadata)

	delimReader := delim_reader.MakeDelimReader('\n', r)
	defer delim_reader.PutDelimReader(delimReader)

	for {
		var line string

		line, err = delimReader.ReadOneString()

		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		trimmed := strings.TrimSpace(line)

		if len(trimmed) == 0 {
			continue
		}

		key, remainder := trimmed[0], strings.TrimSpace(trimmed[1:])

		switch key {
		case '#':
			err = metadata.Description.Set(remainder)

		case '%':
			metadata.Comments = append(metadata.Comments, remainder)

		case '-':
			metadata.AddTagString(remainder)

		case '!':
			err = parser.readType(metadata, remainder)

		default:
			err = errors.ErrorWithStackf("unsupported entry: %q", line)
		}

		if err != nil {
			err = errors.Wrapf(
				err,
				"Line: %q, Key: %q, Value: %q",
				line,
				key,
				remainder,
			)
			return n, err
		}
	}

	return n, err
}

func (parser *textParser2) readType(
	metadata *metadata,
	desc string,
) (err error) {
	if desc == "" {
		return err
	}

	tail := path.Ext(desc)
	head := desc[:len(desc)-len(tail)]

	//! <path>.<typ ext>
	switch {
	case files.Exists(desc):
		if err = metadata.Type.Set(tail); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = parser.Blob.SetWithBlobWriterFactory(desc, parser.BlobWriterFactory); err != nil {
			err = errors.Wrap(err)
			return err
		}

	//! <sha>.<typ ext>
	case tail != "":
		if err = parser.setBlobSha(metadata, head); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = metadata.Type.Set(tail); err != nil {
			err = errors.Wrap(err)
			return err
		}

	//! <sha>
	case tail == "":
		if err = parser.setBlobSha(metadata, head); err == nil {
			return err
		}

		err = nil

		fallthrough

	//! <typ ext>
	default:
		if err = metadata.Type.Set(head); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (parser *textParser2) setBlobSha(
	metadata *metadata,
	maybeSha string,
) (err error) {
	if err = markl.SetMarklIdWithFormatBlech32(
		metadata.GetBlobDigestMutable(),
		markl.PurposeObjectDigestV1,
		maybeSha,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
