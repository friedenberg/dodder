package markl

import (
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

const test = "dodder-repo-private_key-v1@j7putls3hwau0twypl63kpdm3kxvyu4u9dc86692au4muqplslvqmxgey0"

func TestBroken(t1 *testing.T) {
	t1.Skip()
	t := ui.MakeTestContext(t1)

	var id IdBroken

	t.AssertNoError(id.UnmarshalText([]byte(test)))
}
