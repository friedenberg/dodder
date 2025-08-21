package object_metadata

import (
	"io"
	"path"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/delim_reader"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
)

type textParser2 struct {
	interfaces.BlobWriter
	TextParserContext
	Blob fd.FD
}

func (f *textParser2) ReadFrom(r io.Reader) (n int64, err error) {
	m := f.GetMetadata()
	Resetter.Reset(m)

	dr := delim_reader.MakeDelimReader('\n', r)
	defer delim_reader.PutDelimReader(dr)

	for {
		var line string

		line, err = dr.ReadOneString()

		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		trimmed := strings.TrimSpace(line)

		if len(trimmed) == 0 {
			continue
		}

		key, remainder := trimmed[0], strings.TrimSpace(trimmed[1:])

		switch key {
		case '#':
			err = m.Description.Set(remainder)

		case '%':
			m.Comments = append(m.Comments, remainder)

		case '-':
			m.AddTagString(remainder)

		case '!':
			err = f.readTyp(m, remainder)

		default:
			err = errors.ErrorWithStackf("unsupported entry: %q", line)
		}

		if err != nil {
			err = errors.Wrapf(err, "Line: %q, Key: %q, Value: %q", line, key, remainder)
			return
		}
	}

	return
}

func (f *textParser2) readTyp(
	m *Metadata,
	desc string,
) (err error) {
	if desc == "" {
		return
	}

	tail := path.Ext(desc)
	head := desc[:len(desc)-len(tail)]

	//! <path>.<typ ext>
	switch {
	case files.Exists(desc):
		if err = m.Type.Set(tail); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = f.Blob.SetWithBlobWriterFactory(desc, f.BlobWriter); err != nil {
			err = errors.Wrap(err)
			return
		}

	//! <sha>.<typ ext>
	case tail != "":
		if err = f.setBlobSha(m, head); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = m.Type.Set(tail); err != nil {
			err = errors.Wrap(err)
			return
		}

	//! <sha>
	case tail == "":
		if err = f.setBlobSha(m, head); err == nil {
			return
		}

		err = nil

		fallthrough

	//! <typ ext>
	default:
		if err = m.Type.Set(head); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (f *textParser2) setBlobSha(
	m *Metadata,
	maybeSha string,
) (err error) {
	if err = m.Blob.Set(maybeSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
