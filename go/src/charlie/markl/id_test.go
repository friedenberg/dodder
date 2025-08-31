package markl

import (
	"fmt"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestIdNullAndEqual(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	for _, hashType := range hashTypes {
		testIdNullAndEqual(t, hashType)
	}
}

func testIdNullAndEqual(t *ui.TestContext, hashType HashType) {
	{
		t.AssertError(MakeErrIsNull(hashType.null, ""))
		t.AssertNoError(MakeErrIsNotNull(hashType.null))
		t.AssertNoError(
			MakeErrLength(hashType.GetSize(), len(hashType.null.GetBytes())),
		)
	}

	var idZero Id
	hash := hashType.Get()

	{
		idNull, _ := hash.GetMarklId()

		t.AssertNoError(MakeErrIsNotNull(idZero))
		t.AssertNoError(MakeErrIsNotNull(idNull))
		t.AssertError(MakeErrIsNull(idZero, ""))
		t.AssertError(MakeErrIsNull(idNull, ""))
		t.AssertNoError(MakeErrNotEqual(idZero, idNull))
		t.AssertNoError(MakeErrNotEqual(idNull, idZero))
	}

	{
		idNull, _ := hashType.GetMarklIdForString("")

		t.AssertNoError(MakeErrIsNotNull(idZero))
		t.AssertNoError(MakeErrIsNotNull(idNull))
		t.AssertError(MakeErrIsNull(idZero, ""))
		t.AssertError(MakeErrIsNull(idNull, ""))
		t.AssertNoError(MakeErrNotEqual(idZero, idNull))
		t.AssertNoError(MakeErrNotEqual(idNull, idZero))
	}

	{
		idNull, _ := hashType.GetBlobIdForHexString(
			fmt.Sprintf("%x", hashType.null.GetBytes()),
		)

		t.AssertNoError(MakeErrIsNotNull(idZero))
		t.AssertNoError(MakeErrIsNotNull(idNull))
		t.AssertError(MakeErrIsNull(idZero, ""))
		t.AssertError(MakeErrIsNull(idNull, ""))
		t.AssertNoError(MakeErrNotEqual(idZero, idNull))
		t.AssertNoError(MakeErrNotEqual(idNull, idZero))
	}

	{
		idNonZero, _ := hashType.GetMarklIdForString("nonZero")
		t.AssertNoError(MakeErrIsNull(idNonZero, ""))
		t.AssertError(MakeErrIsNotNull(idNonZero))
		t.AssertError(MakeErrNotEqual(idNonZero, idZero))
		t.AssertError(MakeErrNotEqual(idZero, idNonZero))
	}
}

func TestIdEncodeDecode(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	for _, hashType := range hashTypes {
		hash := hashType.Get()

		{
			id, _ := hash.GetMarklId()
			stringValue := StringHRPCombined(id)

			t.Log(stringValue)

			t.AssertNoError(
				SetBlechCombinedHRPAndData(id, stringValue),
			)
		}
	}
}

func FuzzIdStringLen(f *testing.F) {
	f.Add("testing")
	f.Add("holidays")

	hashType := HashTypeBlake2b256

	f.Fuzz(
		func(t1 *testing.T, input string) {
			id, repool := hashType.GetMarklIdForString(input)
			defer repool()

			if input == "" {
				return
			}

			actual := len(id.String()) - len(hashType.GetMarklTypeId())
			expected := 59

			if actual != expected {
				t1.Errorf("expected %d but got %d", expected, actual)
			}
		},
	)
}
