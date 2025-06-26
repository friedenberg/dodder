package dormant_index

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Index struct {
	changes []string
	tags    tag_paths.TagsWithParentsAndTypes
}

func (sch *Index) GetChanges() (out []string) {
	out = make([]string, len(sch.changes))
	copy(out, sch.changes)

	return
}

func (sch *Index) HasChanges() bool {
	return len(sch.changes) > 0
}

func (sch *Index) AddDormantTag(e *tag_paths.Tag) (err error) {
	sch.changes = append(sch.changes, fmt.Sprintf("added %q", e))

	if err = sch.tags.Add(e, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (sch *Index) RemoveDormantTag(e *tag_paths.Tag) (err error) {
	sch.changes = append(sch.changes, fmt.Sprintf("removed %q", e))

	if err = sch.tags.Remove(e); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (sch *Index) ContainsSku(sk *sku.Transacted) bool {
	for _, e := range sch.tags {
		if e.Len() == 0 {
			panic("empty dormant tag")
		}

		all := sk.Metadata.Cache.TagPaths.All
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

func (sch *Index) Load(s env_repo.Env) (err error) {
	var f *os.File

	p := s.FileCacheDormant()

	if f, err = files.Open(p); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	defer errors.DeferredCloser(&err, f)

	br := bufio.NewReader(f)

	if _, err = sch.ReadFrom(br); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (sch *Index) Flush(
	s env_repo.Env,
	printerHeader interfaces.FuncIter[string],
	dryRun bool,
) (err error) {
	if len(sch.changes) == 0 {
		ui.Log().Print("no dormant changes")
		return
	}

	if dryRun {
		ui.Log().Print("no dormant flush, dry run")
		return
	}

	if err = printerHeader("writing dormant"); err != nil {
		err = errors.Wrap(err)
		return
	}

	p := s.FileCacheDormant()

	var f *os.File

	if f, err = files.OpenExclusiveWriteOnlyTruncate(p); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	bw := bufio.NewWriter(f)
	defer errors.DeferredFlusher(&err, bw)

	if _, err = sch.WriteTo(bw); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = printerHeader("wrote dormant"); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s *Index) ReadFrom(r *bufio.Reader) (n int64, err error) {
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

		return
	}

	s.tags = slices.Grow(s.tags, int(count))

	for i := uint16(0); i < count; i++ {
		var l uint16

		var n1 int64
		l, n1, err = ohio.ReadUint16(r)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		var cs *catgut.String

		if cs, err = catgut.MakeFromReader(r, int(l)); err != nil {
			err = errors.Wrap(err)
			return
		}

		s.tags = append(s.tags, tag_paths.TagWithParentsAndTypes{
			Tag: cs,
		})
	}

	return
}

func (s Index) WriteTo(w io.Writer) (n int64, err error) {
	count := uint16(s.tags.Len())

	var n1 int
	var n2 int64
	n1, err = ohio.WriteUint16(w, count)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, e := range s.tags {
		l := uint16(e.Len())

		n1, err = ohio.WriteUint16(w, l)
		n += int64(n1)
		n += int64(n1)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		n2, err = e.WriteTo(w)
		n += n2

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
