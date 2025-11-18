package queries

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/golf/tag_paths"
	"code.linenisgreat.com/dodder/go/src/juliett/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

// TODO move implicit tags here
type Tags struct {
	changes []string
	tags    tag_paths.TagsWithParentsAndTypes
}

func (sch *Tags) GetChanges() (out []string) {
	out = make([]string, len(sch.changes))
	copy(out, sch.changes)

	return out
}

func (sch *Tags) HasChanges() bool {
	return len(sch.changes) > 0
}

func (sch *Tags) AddTag(e *tag_paths.Tag) (err error) {
	sch.changes = append(sch.changes, fmt.Sprintf("added %q", e))

	if err = sch.tags.Add(e, nil); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (sch *Tags) RemoveDormantTag(e *tag_paths.Tag) (err error) {
	sch.changes = append(sch.changes, fmt.Sprintf("removed %q", e))

	if err = sch.tags.Remove(e); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (sch *Tags) ContainsSku(sk *sku.Transacted) bool {
	for _, e := range sch.tags {
		if e.Len() == 0 {
			panic("empty dormant tag")
		}

		all := sk.GetMetadata().GetIndex().GetTagPaths().All
		i, ok := all.ContainsTag(e.Tag)

		if ok {
			ui.Log().Printf(
				"dormant true for %s: %s in %s",
				sk,
				e,
				all[i],
			)

			return true
		}
	}

	ui.Log().Printf(
		"dormant false for %s",
		sk,
	)

	return false
}

func (sch *Tags) Load(s env_repo.Env) (err error) {
	var f *os.File

	p := s.FileTags()

	if f, err = files.Open(p); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	defer errors.DeferredCloser(&err, f)

	br := bufio.NewReader(f)

	if _, err = sch.ReadFrom(br); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (sch *Tags) Flush(
	s env_repo.Env,
	printerHeader interfaces.FuncIter[string],
	dryRun bool,
) (err error) {
	if len(sch.changes) == 0 {
		ui.Log().Print("no tags changes")
		return err
	}

	if dryRun {
		ui.Log().Print("no tags flush, dry run")
		return err
	}

	if err = printerHeader("writing dormant tags"); err != nil {
		err = errors.Wrap(err)
		return err
	}

	p := s.FileTags()

	var f *os.File

	if f, err = files.OpenExclusiveWriteOnlyTruncate(p); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, f)

	bw := bufio.NewWriter(f)
	defer errors.DeferredFlusher(&err, bw)

	if _, err = sch.WriteTo(bw); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = printerHeader("wrote dormant"); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (s *Tags) ReadFrom(r *bufio.Reader) (n int64, err error) {
	s.tags.Reset()
	var count uint16

	var n1 int64
	count, n1, err = ohio.ReadUint16(r)
	n += n1
	// n += int64(n1)
	if err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return n, err
	}

	s.tags = slices.Grow(s.tags, int(count))

	for i := uint16(0); i < count; i++ {
		var l uint16

		var n1 int64
		l, n1, err = ohio.ReadUint16(r)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		var cs *catgut.String

		if cs, err = catgut.MakeFromReader(r, int(l)); err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		s.tags = append(s.tags, tag_paths.TagWithParentsAndTypes{
			Tag: cs,
		})
	}

	return n, err
}

func (s Tags) WriteTo(w io.Writer) (n int64, err error) {
	count := uint16(s.tags.Len())

	var n1 int
	var n2 int64
	n1, err = ohio.WriteUint16(w, count)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	for _, e := range s.tags {
		l := uint16(e.Len())

		n1, err = ohio.WriteUint16(w, l)
		n += int64(n1)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		n2, err = e.WriteTo(w)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}
	}

	return n, err
}
