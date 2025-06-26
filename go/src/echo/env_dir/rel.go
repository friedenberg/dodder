package env_dir

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type RelativePath interface {
	Rel(string) string
}

func (s env) MakeRelativePathStringFormatWriter() interfaces.StringEncoderTo[string] {
	return relativePathStringFormatWriter(s)
}

type relativePathStringFormatWriter env

func (f relativePathStringFormatWriter) EncodeStringTo(
	p string,
	w interfaces.WriterAndStringWriter,
) (n int64, err error) {
	var n1 int

	{
		// if p, err = filepath.Rel(s.cwd, p); err != nil {
		// 	err = errors.Wrap(err)
		// 	return
		// }

		p1, _ := filepath.Rel(f.cwd, p)

		if p1 != "" {
			p = p1
		}
	}

	n1, err = w.WriteString(p)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
