package sku

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/delta/thyme"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

// TODO rename to FS
type FSItem struct {
	// TODO refactor this to be a string and a genre that is tied to the state
	ExternalObjectId ids.ExternalObjectId

	Object   fd.FD
	Blob     fd.FD // TODO make set
	Conflict fd.FD

	interfaces.MutableSetLike[*fd.FD]
}

func (ef *FSItem) WriteToSku(
	external *Transacted,
	dirLayout env_dir.Env,
) (err error) {
	if err = ef.WriteToExternalObjectId(
		&external.ExternalObjectId,
		dirLayout,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (ef *FSItem) WriteToExternalObjectId(
	eoid *ids.ExternalObjectId,
	dirLayout env_dir.Env,
) (err error) {
	eoid.SetGenre(ef.ExternalObjectId.GetGenre())

	var relPath string
	var anchorFD *fd.FD

	switch {
	case !ef.Object.IsEmpty():
		anchorFD = &ef.Object

	case !ef.Blob.IsEmpty():
		anchorFD = &ef.Blob

	case !ef.Conflict.IsEmpty():
		anchorFD = &ef.Conflict

	default:
		// [int/tanz @0a9d !task project-2021-zit-bugs zz-inbox] fix nil pointer during organize in workspace
		ui.Err().Printf("item has no anchor FDs. %q", ef.Debug())
		return
	}

	relPath = dirLayout.RelToCwdOrSame(anchorFD.GetPath())

	if relPath == "-" {
		return
	}

	if err = eoid.Set(relPath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (ef *FSItem) String() string {
	return ef.ExternalObjectId.String()
}

func (ef *FSItem) GetExternalObjectId() *ids.ExternalObjectId {
	return &ef.ExternalObjectId
}

func (i *FSItem) Debug() string {
	return fmt.Sprintf(
		"Genre: %q, ObjectId: %q, Object: %q, Blob: %q, Conflict: %q, All: %q",
		i.ExternalObjectId.GetGenre(),
		&i.ExternalObjectId,
		&i.Object,
		&i.Blob,
		&i.Conflict,
		i.MutableSetLike,
	)
}

func (i *FSItem) GetTai() ids.Tai {
	return ids.TaiFromTime(i.LatestModTime())
}

func (i *FSItem) GetTime() thyme.Time {
	return i.LatestModTime()
}

func (i *FSItem) LatestModTime() thyme.Time {
	o, b := i.Object.ModTime(), i.Blob.ModTime()

	if o.Less(b) {
		return b
	} else {
		return o
	}
}

func (dst *FSItem) Reset() {
	dst.ExternalObjectId.Reset()
	dst.Object.Reset()
	dst.Blob.Reset()
	dst.Conflict.Reset()

	if dst.MutableSetLike == nil {
		dst.MutableSetLike = collections_value.MakeMutableValueSet[*fd.FD](nil)
	} else {
		dst.MutableSetLike.Reset()
	}
}

func (dst *FSItem) ResetWith(src *FSItem) {
	if dst == src {
		return
	}

	dst.ExternalObjectId.ResetWith(&src.ExternalObjectId)
	dst.Object.ResetWith(&src.Object)
	dst.Blob.ResetWith(&src.Blob)
	dst.Conflict.ResetWith(&src.Conflict)

	if dst.MutableSetLike == nil {
		dst.MutableSetLike = collections_value.MakeMutableValueSet[*fd.FD](nil)
	}

	dst.MutableSetLike.Reset()

	if src.MutableSetLike != nil {
		for item := range src.MutableSetLike.All() {
			dst.MutableSetLike.Add(item)
		}
	}

	// TODO consider if this approach actually works
	if !dst.Object.IsEmpty() {
		dst.Add(&dst.Object)
	}

	if !dst.Blob.IsEmpty() {
		dst.Add(&dst.Blob)
	}

	if !dst.Conflict.IsEmpty() {
		dst.Add(&dst.Conflict)
	}
}

func (a *FSItem) Equals(b *FSItem) (ok bool, why string) {
	if ok, why = a.Object.Equals2(&b.Object); !ok {
		return false, fmt.Sprintf("Object.%s", why)
	}

	if ok, why = a.Blob.Equals2(&b.Blob); !ok {
		return false, fmt.Sprintf("Blob.%s", why)
	}

	if ok, why = a.Conflict.Equals2(&b.Conflict); !ok {
		return false, fmt.Sprintf("Conflict.%s", why)
	}

	if !quiter.SetEquals(a.MutableSetLike, b.MutableSetLike) {
		return false, "set"
	}

	return
}

func (e *FSItem) GenerateConflictFD() (err error) {
	if e.ExternalObjectId.IsEmpty() {
		err = errors.ErrorWithStackf("cannot generate conflict FD for empty external object id")
		return
	}

	if err = e.Conflict.SetPath(e.ExternalObjectId.String() + ".conflict"); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (e *FSItem) GetCheckoutModeOrError() (m checkout_mode.Mode, err error) {
	switch {
	case !e.Object.IsEmpty() && !e.Blob.IsEmpty():
		m = checkout_mode.MetadataAndBlob

	case !e.Blob.IsEmpty():
		m = checkout_mode.BlobOnly

	case !e.Object.IsEmpty():
		m = checkout_mode.MetadataOnly

	case !e.Conflict.IsEmpty():
		err = MakeErrMergeConflict(e)

	default:
		err = checkout_mode.MakeErrInvalidCheckoutMode(
			errors.ErrorWithStackf("all FD's are empty: %s", e.Debug()),
		)
	}

	return
}

func (e *FSItem) GetCheckoutMode() (m checkout_mode.Mode) {
	switch {
	case !e.Object.IsEmpty() && !e.Blob.IsEmpty():
		m = checkout_mode.MetadataAndBlob

	case !e.Blob.IsEmpty():
		m = checkout_mode.BlobOnly

	case !e.Object.IsEmpty():
		m = checkout_mode.MetadataOnly
	}

	return
}
