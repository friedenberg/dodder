package sha

import (
	"testing"
	"unsafe"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestSize(t1 *testing.T) {
	t1.Skip()
	t := ui.MakeTestContext(t1)

	type testStruct struct {
		data [ByteSize + 1]byte
	}

	type testStruct2 struct {
		nonZero bool
		data    [ByteSize]byte
	}

	{
		size := unsafe.Sizeof(testStruct{})

		if size != 0 {
			t.Errorf("size: %d", size)
		}
	}

	{
		size := unsafe.Sizeof(testStruct2{})

		if size != 0 {
			t.Errorf("size: %d", size)
		}
	}
}
