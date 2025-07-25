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

	FDs interfaces.MutableSetLike[*fd.FD]
}

func (item *FSItem) WriteToSku(
	external *Transacted,
	dirLayout env_dir.Env,
) (err error) {
	if err = item.WriteToExternalObjectId(
		&external.ExternalObjectId,
		dirLayout,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (item *FSItem) WriteToExternalObjectId(
	eoid *ids.ExternalObjectId,
	dirLayout env_dir.Env,
) (err error) {
	eoid.SetGenre(item.ExternalObjectId.GetGenre())

	var relPath string
	var anchorFD *fd.FD

	switch {
	case !item.Object.IsEmpty():
		anchorFD = &item.Object

	case !item.Blob.IsEmpty():
		anchorFD = &item.Blob

	case !item.Conflict.IsEmpty():
		anchorFD = &item.Conflict

	default:
		// [int/tanz @0a9d !task project-2021-zit-bugs zz-inbox] fix nil pointer
		// during organize in workspace
		ui.Err().Printf("item has no anchor FDs. %q", item.Debug())
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

func (item *FSItem) String() string {
	return item.ExternalObjectId.String()
}

func (item *FSItem) GetExternalObjectId() *ids.ExternalObjectId {
	return &item.ExternalObjectId
}

func (item *FSItem) Debug() string {
	return fmt.Sprintf(
		"Genre: %q, ObjectId: %q, Object: %q, Blob: %q, Conflict: %q, All: %q",
		item.ExternalObjectId.GetGenre(),
		&item.ExternalObjectId,
		&item.Object,
		&item.Blob,
		&item.Conflict,
		item.FDs,
	)
}

func (item *FSItem) GetTai() ids.Tai {
	return ids.TaiFromTime(item.LatestModTime())
}

func (item *FSItem) GetTime() thyme.Time {
	return item.LatestModTime()
}

func (item *FSItem) LatestModTime() thyme.Time {
	o, b := item.Object.ModTime(), item.Blob.ModTime()

	if o.Less(b) {
		return b
	} else {
		return o
	}
}

func (item *FSItem) Reset() {
	item.ExternalObjectId.Reset()
	item.Object.Reset()
	item.Blob.Reset()
	item.Conflict.Reset()

	if item.FDs == nil {
		item.FDs = collections_value.MakeMutableValueSet[*fd.FD](nil)
	} else {
		item.FDs.Reset()
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

	if dst.FDs == nil {
		dst.FDs = collections_value.MakeMutableValueSet[*fd.FD](nil)
	}

	dst.FDs.Reset()

	if src.FDs != nil {
		for item := range src.FDs.All() {
			dst.FDs.Add(item)
		}
	}

	// TODO consider if this approach actually works
	if !dst.Object.IsEmpty() {
		dst.FDs.Add(&dst.Object)
	}

	if !dst.Blob.IsEmpty() {
		dst.FDs.Add(&dst.Blob)
	}

	if !dst.Conflict.IsEmpty() {
		dst.FDs.Add(&dst.Conflict)
	}
}

func (item *FSItem) Equals(b *FSItem) (ok bool, why string) {
	if ok, why = item.Object.Equals2(&b.Object); !ok {
		return false, fmt.Sprintf("Object.%s", why)
	}

	if ok, why = item.Blob.Equals2(&b.Blob); !ok {
		return false, fmt.Sprintf("Blob.%s", why)
	}

	if ok, why = item.Conflict.Equals2(&b.Conflict); !ok {
		return false, fmt.Sprintf("Conflict.%s", why)
	}

	if !quiter.SetEquals(item.FDs, b.FDs) {
		return false, "set"
	}

	return
}

func (item *FSItem) GenerateConflictFD() (err error) {
	if item.ExternalObjectId.IsEmpty() {
		err = errors.ErrorWithStackf(
			"cannot generate conflict FD for empty external object id",
		)
		return
	}

	if err = item.Conflict.SetPath(item.ExternalObjectId.String() + ".conflict"); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (item *FSItem) GetCheckoutModeOrError() (m checkout_mode.Mode, err error) {
	switch {
	case !item.Object.IsEmpty() && !item.Blob.IsEmpty():
		m = checkout_mode.MetadataAndBlob

	case !item.Blob.IsEmpty():
		m = checkout_mode.BlobOnly

	case !item.Object.IsEmpty():
		m = checkout_mode.MetadataOnly

	case !item.Conflict.IsEmpty():
		err = MakeErrMergeConflict(item)

	default:
		err = checkout_mode.MakeErrInvalidCheckoutMode(
			errors.ErrorWithStackf("all FD's are empty: %s", item.Debug()),
		)
	}

	return
}

func (item *FSItem) GetCheckoutMode() (m checkout_mode.Mode) {
	switch {
	case !item.Object.IsEmpty() && !item.Blob.IsEmpty():
		m = checkout_mode.MetadataAndBlob

	case !item.Blob.IsEmpty():
		m = checkout_mode.BlobOnly

	case !item.Object.IsEmpty():
		m = checkout_mode.MetadataOnly
	}

	return
}
