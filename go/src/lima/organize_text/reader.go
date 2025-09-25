package organize_text

import (
	"io"
	"unicode"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/unicorn"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/id_fmts"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type reader struct {
	options           Options
	lineNo            int
	root              *Assignment
	currentAssignment *Assignment
}

func (ar *reader) ReadFrom(r1 io.Reader) (n int64, err error) {
	r := catgut.MakeRingBuffer(r1, 0)
	rbs := catgut.MakeRingBufferScanner(r)

	ar.root = newAssignment(0)
	ar.currentAssignment = ar.root

LOOP:
	for {
		var sl catgut.Slice
		var offsetPlusMatch int

		sl, offsetPlusMatch, err = rbs.FirstMatch(unicorn.Not(unicode.IsSpace))

		if err == io.EOF && sl.Len() == 0 {
			err = nil
			break
		}

		switch err {
		case catgut.ErrBufferEmpty, catgut.ErrNoMatch:
			var n1 int64
			n1, err = r.Fill()

			if n1 == 0 && err == io.EOF {
				err = nil
				break LOOP
			} else {
				err = nil
				continue
			}
		}

		if err != nil && err != io.EOF {
			err = errors.Wrap(err)
			return n, err
		}

		r.AdvanceRead(offsetPlusMatch)
		n += int64(sl.Len())
		sb := sl.SliceBytes()

		slen := sl.Len()

		if slen >= 1 {
			pr := sl.FirstByte()

			switch pr {
			case '#':
				if err = ar.readOneHeading(r, sb); err != nil {
					err = errors.Wrap(err)
					return n, err
				}

			case '%':
				if err = ar.readOneObj(r, tag_paths.TypeUnknown); err != nil {
					if err == io.EOF {
						err = nil
					} else {
						err = errors.Wrap(err)
						return n, err
					}
				}

			case '-':
				if err = ar.readOneObj(r, tag_paths.TypeDirect); err != nil {
					if err == io.EOF {
						err = nil
					} else {
						err = errors.Wrap(err)
						return n, err
					}
				}

			default:
				err = errors.ErrorWithStackf("unsupported verb: %c. slice: %q", pr, sl)
				return n, err
			}
		}

		ar.lineNo++

		if err == io.EOF {
			err = nil
			break
		} else {
			continue
		}
	}

	return n, err
}

func (ar *reader) readOneHeading(
	rb *catgut.RingBuffer,
	match catgut.SliceBytes,
) (err error) {
	depth := unicorn.CountRune(match.Bytes, '#')

	currentTags := ids.MakeMutableTagSet()

	reader := id_fmts.MakeTagsReader()

	if _, err = reader.ReadStringFormat(currentTags, rb); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var newAssignment *Assignment

	if depth < ar.currentAssignment.Depth {
		newAssignment, err = ar.readOneHeadingLesserDepth(
			depth,
			currentTags,
		)
	} else if depth == ar.currentAssignment.Depth {
		newAssignment, err = ar.readOneHeadingEqualDepth(depth, currentTags)
	} else {
		// always use currentTags.depth + 1 because it corrects movements
		newAssignment, err = ar.readOneHeadingGreaterDepth(depth, currentTags)
	}

	if err != nil {
		err = ErrorRead{
			error:  err,
			line:   ar.lineNo,
			column: 2,
		}

		return err
	}

	if newAssignment == nil {
		err = errors.ErrorWithStackf("read heading function return nil new assignment")
		return err
	}

	ar.currentAssignment = newAssignment

	return err
}

func (ar *reader) readOneHeadingLesserDepth(
	d int,
	e ids.TagSet,
) (newCurrent *Assignment, err error) {
	depthDiff := d - ar.currentAssignment.GetDepth()

	if newCurrent, err = ar.currentAssignment.nthParent(depthDiff - 1); err != nil {
		err = errors.Wrap(err)
		return newCurrent, err
	}

	if e.Len() == 0 {
		// `
		// # task-todo
		// ## priority-1
		// - wow
		// #
		// `
		// logz.Print("new set is empty")
	} else {
		// `
		// # task-todo
		// ## priority-1
		// - wow
		// # zz-inbox
		// `
		assignment := newAssignment(d)
		assignment.Transacted.Metadata.Tags = e.CloneMutableSetPtrLike()
		newCurrent.addChild(assignment)
		newCurrent = assignment
	}

	return newCurrent, err
}

func (ar *reader) readOneHeadingEqualDepth(
	d int,
	e ids.TagSet,
) (newCurrent *Assignment, err error) {
	// logz.Print("depth count is ==")

	if newCurrent, err = ar.currentAssignment.nthParent(1); err != nil {
		err = errors.Wrap(err)
		return newCurrent, err
	}

	if e.Len() == 0 {
		// `
		// # task-todo
		// ## priority-1
		// - wow
		// ##
		// `
		// logz.Print("new set is empty")
	} else {
		// `
		// # task-todo
		// ## priority-1
		// - wow
		// ## priority-2
		// `
		assignment := newAssignment(d)
		assignment.Transacted.Metadata.Tags = e.CloneMutableSetPtrLike()
		newCurrent.addChild(assignment)
		newCurrent = assignment
	}

	return newCurrent, err
}

func (ar *reader) readOneHeadingGreaterDepth(
	d int,
	e ids.TagSet,
) (newCurrent *Assignment, err error) {
	// logz.Print("depth count is >")
	// logz.Print(e)

	newCurrent = ar.currentAssignment

	if e.Len() == 0 {
		// `
		// # task-todo
		// ## priority-1
		// - wow
		// ###
		// `
		// logz.Print("new set is empty")
	} else {
		// `
		// # task-todo
		// ## priority-1
		// - wow
		// ### priority-2
		// `
		assignment := newAssignment(d)
		assignment.Transacted.Metadata.Tags = e.CloneMutableSetPtrLike()
		newCurrent.addChild(assignment)
		// logz.Print("adding to parent")
		// logz.Print("child", assignment)
		// logz.Print("parent", newCurrent)
		newCurrent = assignment
	}

	return newCurrent, err
}

func (ar *reader) readOneObj(
	r *catgut.RingBuffer,
	t tag_paths.Type,
) (err error) {
	// logz.Print("reading one zettel", l)

	var z obj
	z.sku = ar.options.ObjectFactory.Get()
	z.tipe = t

	if _, err = ar.options.fmtBox.ReadStringFormat(
		z.GetSkuExternal(),
		catgut.MakeRingBufferRuneScanner(r),
	); err != nil {
		err = ErrorRead{
			error:  err,
			line:   ar.lineNo,
			column: 2,
		}

		return err
	}

	// z.External.GetSkuExternal().Metadata.Tai = ids.NowTai()

	// if err = z.External.GetSkuExternal().CalculateObjectShas(); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	if z.GetSkuExternal().ObjectId.IsEmpty() {
		z.sku.SetState(checked_out_state.Untracked)

		// set empty zettel id to ensure middle is '/'
		if err = z.GetSkuExternal().ObjectId.SetWithIdLike(ids.ZettelId{}); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else {
		z.sku.SetState(checked_out_state.CheckedOut)

		if err = ar.options.Abbr.ExpandZettelIdOnly(&z.GetSkuExternal().ObjectId); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	sku.TransactedResetter.ResetWith(z.GetSku(), z.GetSkuExternal())
	ar.currentAssignment.AddObject(&z)

	return err
}
