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
		t.AssertError(AssertIdIsNotNull(hashType.null, ""))
		t.AssertNoError(AssertIdIsNull(hashType.null))
		t.AssertNoError(
			MakeErrLength(hashType.GetSize(), len(hashType.null.GetBytes())),
		)
	}

	var idZero Id
	hash := hashType.Get()

	{
		idNull, _ := hash.GetMarklId()

		t.AssertNoError(AssertIdIsNull(idZero))
		t.AssertNoError(AssertIdIsNull(idNull))
		t.AssertError(AssertIdIsNotNull(idZero, ""))
		t.AssertError(AssertIdIsNotNull(idNull, ""))
		t.AssertNoError(AssertEqual(idZero, idNull))
		t.AssertNoError(AssertEqual(idNull, idZero))
	}

	{
		idNull, _ := hashType.GetMarklIdForString("")

		t.AssertNoError(AssertIdIsNull(idZero))
		t.AssertNoError(AssertIdIsNull(idNull))
		t.AssertError(AssertIdIsNotNull(idZero, ""))
		t.AssertError(AssertIdIsNotNull(idNull, ""))
		t.AssertNoError(AssertEqual(idZero, idNull))
		t.AssertNoError(AssertEqual(idNull, idZero))
	}

	{
		idNull, _ := hashType.GetBlobIdForHexString(
			fmt.Sprintf("%x", hashType.null.GetBytes()),
		)

		t.AssertNoError(AssertIdIsNull(idZero))
		t.AssertNoError(AssertIdIsNull(idNull))
		t.AssertError(AssertIdIsNotNull(idZero, ""))
		t.AssertError(AssertIdIsNotNull(idNull, ""))
		t.AssertNoError(AssertEqual(idZero, idNull))
		t.AssertNoError(AssertEqual(idNull, idZero))
	}

	{
		idNonZero, _ := hashType.GetMarklIdForString("nonZero")
		t.AssertNoError(AssertIdIsNotNull(idNonZero, ""))
		t.AssertError(AssertIdIsNull(idNonZero))
		t.AssertError(AssertEqual(idNonZero, idZero))
		t.AssertError(AssertEqual(idZero, idNonZero))
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
