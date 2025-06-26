package local_working_copy

import (
	"fmt"
	"maps"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	FormatFunc func(*Repo, interfaces.WriterAndStringWriter) (interfaces.FuncIter[*sku.Transacted], error)

	Format struct {
		*Repo
		value  string
		format interfaces.FuncIter[*sku.Transacted]
	}
)

func (f *Format) Set(v string) (err error) {
	var ok bool
	var rawFormatter FormatFunc

	if rawFormatter, ok = formatters[v]; !ok {
		err = errors.BadRequestf(
			"unsupported format: %q. Available formats: %q",
			v,
			slices.Collect(maps.Keys(formatters)),
		)

		return
	}

	f.value = v

	if f.format, err = rawFormatter(f.Repo, f.Repo.GetOutFile()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (f *Format) String() string {
	if f == nil || f.format == nil {
		return fmt.Sprintf(
			"%q",
			slices.Collect(maps.Keys(formatters)),
		)
	} else {
		return f.value
	}
}

func (u *Format) Format(
	sk *sku.Transacted,
) (err error) {
	if u.format == nil {
	}

	if err = u.format(sk); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

var formatters = map[string]FormatFunc{
	"tags-path": func(u *Repo, out interfaces.WriterAndStringWriter) (f interfaces.FuncIter[*sku.Transacted], err error) {
		f = func(tl *sku.Transacted) (err error) {
			if _, err = fmt.Fprintln(
				out,
				tl.GetObjectId(),
				&tl.Metadata.Cache.TagPaths,
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}

		return
	},
}
