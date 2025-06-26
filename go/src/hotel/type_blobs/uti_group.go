package type_blobs

import "code.linenisgreat.com/dodder/go/src/bravo/equality"

type UTIGroup map[string]string

func (a UTIGroup) Map() map[string]string {
	return map[string]string(a)
}

func (a *UTIGroup) Equals(b UTIGroup) bool {
	if b == nil {
		return false
	}

	if len(a.Map()) != len(b.Map()) {
		return false
	}

	if !equality.MapsOrdered(a.Map(), b.Map()) {
		return false
	}

	return true
}

func (ct *UTIGroup) Merge(ct2 UTIGroup) {
	for k, v := range ct2.Map() {
		if v != "" {
			ct.Map()[k] = v
		}
	}
}
