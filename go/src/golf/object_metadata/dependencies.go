package object_metadata

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/format"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type Dependencies struct {
	EnvDir        env_dir.Env
	BlobStore     interfaces.BlobStore
	BlobFormatter script_config.RemoteScript
}

func (deps Dependencies) GetBlobDigestType() interfaces.HashType {
	hashType := deps.BlobStore.GetDefaultHashType()
	if hashType == nil {
		return markl.HashTypeSha256
	} else {
		return hashType
	}
}

func (deps Dependencies) writeComments(
	writer io.Writer,
	context TextFormatterContext,
) (n int64, err error) {
	n1 := 0

	for _, c := range context.GetMetadata().Comments {
		n1, err = io.WriteString(writer, "% ")
		n += int64(n1)

		if err != nil {
			return
		}

		n1, err = io.WriteString(writer, c)
		n += int64(n1)

		if err != nil {
			return
		}

		n1, err = io.WriteString(writer, "\n")
		n += int64(n1)

		if err != nil {
			return
		}
	}

	return
}

func (deps Dependencies) writeBoundary(
	w1 io.Writer,
	_ TextFormatterContext,
) (n int64, err error) {
	return ohio.WriteLine(w1, triple_hyphen_io.Boundary)
}

func (deps Dependencies) writeNewLine(
	w1 io.Writer,
	_ TextFormatterContext,
) (n int64, err error) {
	return ohio.WriteLine(w1, "")
}

func (deps Dependencies) writeCommonMetadataFormat(
	w1 io.Writer,
	c TextFormatterContext,
) (n int64, err error) {
	w := format.NewLineWriter()
	m := c.GetMetadata()

	if m.Description.String() != "" || !c.DoNotWriteEmptyDescription {
		reader, repool := pool.GetStringReader(m.Description.String())
		defer repool()
		sr := bufio.NewReader(reader)

		for {
			var line string
			line, err = sr.ReadString('\n')
			isEOF := err == io.EOF

			if err != nil && !isEOF {
				err = errors.Wrap(err)
				return
			}

			w.WriteLines(
				fmt.Sprintf("# %s", strings.TrimSpace(line)),
			)

			if isEOF {
				break
			}
		}
	}

	for _, e := range quiter.SortedValues(m.GetTags()) {
		if ids.IsEmpty(e) {
			continue
		}

		w.WriteFormat("- %s", e)
	}

	if n, err = w.WriteTo(w1); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (deps Dependencies) writeTyp(
	w1 io.Writer,
	c TextFormatterContext,
) (n int64, err error) {
	m := c.GetMetadata()

	if m.Type.IsEmpty() {
		return
	}

	return ohio.WriteLine(w1, fmt.Sprintf("! %s", m.Type.StringSansOp()))
}

func (deps Dependencies) writeBlobDigestAndType(
	w1 io.Writer,
	c TextFormatterContext,
) (n int64, err error) {
	m := c.GetMetadata()
	return ohio.WriteLine(
		w1,
		fmt.Sprintf(
			"! %s.%s",
			&m.DigBlob,
			m.Type.StringSansOp(),
		),
	)
}

func (deps Dependencies) writePathType(
	w1 io.Writer,
	c TextFormatterContext,
) (n int64, err error) {
	var ap string

	for _, f := range c.PersistentFormatterContext.GetMetadata().Fields {
		if strings.ToLower(f.Key) == "blob" {
			ap = f.Value
			break
		}
	}

	if ap != "" {
		ap = deps.EnvDir.RelToCwdOrSame(ap)
	} else {
		err = errors.ErrorWithStackf("path not found in fields")
		return
	}

	return ohio.WriteLine(w1, fmt.Sprintf("! %s", ap))
}

func (deps Dependencies) writeBlob(
	w1 io.Writer,
	c TextFormatterContext,
) (n int64, err error) {
	var ar io.ReadCloser
	m := c.GetMetadata()

	if ar, err = deps.BlobStore.BlobReader(&m.DigBlob); err != nil {
		err = errors.Wrap(err)
		return
	}

	if ar == nil {
		err = errors.ErrorWithStackf("blob reader is nil")
		return
	}

	defer errors.DeferredCloser(&err, ar)

	if deps.BlobFormatter != nil {
		var wt io.WriterTo

		if wt, err = script_config.MakeWriterToWithStdin(
			deps.BlobFormatter,
			deps.EnvDir.MakeCommonEnv(),
			ar,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if n, err = wt.WriteTo(w1); err != nil {
			var errExit *exec.ExitError

			if errors.As(err, &errExit) {
				err = MakeErrBlobFormatterFailed(errExit)
			}

			err = errors.Wrap(err)

			return
		}
	} else {
		if n, err = io.Copy(w1, ar); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
