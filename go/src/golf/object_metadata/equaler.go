package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/bravo/blob_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

var (
	EqualerSansTai               equaler
	Equaler                      = equaler{includeTai: true}
	EqualerSansTaiIncludeVirtual = equaler{includeVirtual: true}
)

type equaler struct {
	includeVirtual bool
	includeTai     bool
}

const debug = false

// TODO make better diffing facility
func (e equaler) Equals(a, b *Metadata) bool {
	if e.includeTai && !a.Tai.Equals(b.Tai) {
		if debug {
			ui.Debug().Print(&a.Tai, "->", &b.Tai)
		}
		return false
	}

	if !blob_ids.Equals(&a.Blob, &b.Blob) {
		if debug {
			ui.Debug().Print(&a.Blob, "->", &b.Blob)
		}
		return false
	}

	if !a.Type.Equals(b.Type) {
		if debug {
			ui.Debug().Print(&a.Type, "->", &b.Type)
		}
		return false
	}

	aes := a.GetTags()
	bes := b.GetTags()

	found := false
	for ea := range aes.AllPtr() {
		if (!e.includeVirtual && ea.IsVirtual()) || ea.IsEmpty() {
			continue
		}

		if !bes.ContainsKey(bes.KeyPtr(ea)) {
			if debug {
				ui.Debug().Print(ea, "-> X")
			}
			found = true
			break
		}
	}
	if found {
		if debug {
			ui.Debug().Print(aes, "->", bes)
		}

		return false
	}

	found2 := false
	for eb := range bes.AllPtr() {
		if !e.includeVirtual && eb.IsVirtual() {
			continue
		}

		if !aes.ContainsKey(aes.KeyPtr(eb)) {
			found2 = true
			break
		}
	}
	if found2 {
		if debug {
			ui.Debug().Print(aes, "->", bes)
		}
		return false
	}

	if !a.Description.Equals(b.Description) {
		if debug {
			ui.Debug().Print(a.Description, "->", b.Description)
		}
		return false
	}

	return true
}
