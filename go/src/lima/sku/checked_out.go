package sku

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/external_state"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
)

func InternalAndExternalEqualsWithoutTai(co SkuType) bool {
	i := co.GetSku()
	e := co.GetSkuExternal().GetSku()

	return object_metadata.EqualerSansTai.Equals(&e.Metadata, &i.Metadata)
}

type CheckedOut struct {
	internal Transacted
	external Transacted
	state    checked_out_state.State
}

func (c *CheckedOut) GetRepoId() ids.RepoId {
	return c.GetSkuExternal().RepoId
}

func (c *CheckedOut) GetSkuExternal() *Transacted {
	return &c.external
}

func (c *CheckedOut) GetSku() *Transacted {
	return &c.internal
}

func (c *CheckedOut) GetState() checked_out_state.State {
	return c.state
}

func (src *CheckedOut) Clone() *CheckedOut {
	dst := GetCheckedOutPool().Get()
	CheckedOutResetter.ResetWith(dst, src)
	return dst
}

func (t *CheckedOut) GetExternalObjectId() ids.ExternalObjectIdLike {
	return t.GetSkuExternal().GetExternalObjectId()
}

func (t *CheckedOut) GetExternalState() external_state.State {
	return t.GetSkuExternal().GetExternalState()
}

func (a *CheckedOut) GetObjectId() *ids.ObjectId {
	return a.GetSkuExternal().GetObjectId()
}

func (c *CheckedOut) SetState(v checked_out_state.State) (err error) {
	c.state = v
	return err
}

func (a *CheckedOut) String() string {
	return fmt.Sprintf("%s %s", a.GetSku(), a.GetSkuExternal())
}

func (a *CheckedOut) Equals(b *CheckedOut) bool {
	return a.internal.Equals(&b.internal) && a.external.Equals(&b.external)
}

func (a *CheckedOut) GetTai() ids.Tai {
	external := a.external.GetTai()

	if external.IsZero() {
		return a.internal.GetTai()
	} else {
		return external
	}
}
