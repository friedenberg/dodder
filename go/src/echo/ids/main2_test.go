package ids

import (
	"bytes"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func idWriteToReadFromData() []string {
	return []string{
		"one/uno",
		"config",
		"!md",
		"-tag",
		"//repo",
		"tag",
	}
}

func TestIdWriteToReadFrom(t1 *testing.T) {
	t := ui.T{T: t1}
	for _, v := range idWriteToReadFromData() {
		var k ObjectId
		t.AssertNoError(k.Set(v))

		var b bytes.Buffer

		_, err := k.WriteTo(&b)
		t.AssertNoError(err)

		var k2 ObjectId

		_, err = k2.ReadFrom(&b)
		t.AssertNoError(err)

		if k.String() != k2.String() {
			t.NotEqual(&k, &k2)
		}
	}
}
