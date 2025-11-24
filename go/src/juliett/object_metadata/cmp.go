package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
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
func (e equaler) Equals(a, b IMetadata) bool {
	{
		a := a.(*metadata)
		b := b.(*metadata)

		if e.includeTai && !a.Tai.Equals(b.Tai) {
			if debug {
				ui.Debug().Print(&a.Tai, "->", &b.Tai)
			}
			return false
		}

		if !markl.Equals(&a.DigBlob, &b.DigBlob) {
			if debug {
				ui.Debug().Print(&a.DigBlob, "->", &b.DigBlob)
			}
			return false
		}

		if !a.GetType().Equals(b.GetType()) {
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
}

var Lessor lessor

type lessor struct{}

func (lessor) Less(a, b *metadata) bool {
	return a.Tai.Less(b.Tai)
}
