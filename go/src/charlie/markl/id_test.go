package markl

import (
	"fmt"
	"testing"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

func TestIdNullAndEqual(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	for _, formatHash := range formatHashes {
		testIdNullAndEqual(t, formatHash)
	}
}

func testIdNullAndEqual(t *ui.TestContext, formatHash FormatHash) {
	{
		t.AssertError(AssertIdIsNotNull(formatHash.null, ""))
		t.AssertNoError(AssertIdIsNull(formatHash.null))
		t.AssertNoError(
			MakeErrLength(
				formatHash.GetSize(),
				len(formatHash.null.GetBytes()),
			),
		)
	}

	var idZero Id
	hash := formatHash.Get()

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
		idNull, _ := formatHash.GetMarklIdForString("")

		t.AssertNoError(AssertIdIsNull(idZero))
		t.AssertNoError(AssertIdIsNull(idNull))
		t.AssertError(AssertIdIsNotNull(idZero, ""))
		t.AssertError(AssertIdIsNotNull(idNull, ""))
		t.AssertNoError(AssertEqual(idZero, idNull))
		t.AssertNoError(AssertEqual(idNull, idZero))
	}

	{
		idNull, _ := formatHash.GetBlobIdForHexString(
			fmt.Sprintf("%x", formatHash.null.GetBytes()),
		)

		t.AssertNoError(AssertIdIsNull(idZero))
		t.AssertNoError(AssertIdIsNull(idNull))
		t.AssertError(AssertIdIsNotNull(idZero, ""))
		t.AssertError(AssertIdIsNotNull(idNull, ""))
		t.AssertNoError(AssertEqual(idZero, idNull))
		t.AssertNoError(AssertEqual(idNull, idZero))
	}

	{
		idNonZero, _ := formatHash.GetMarklIdForString("nonZero")
		t.AssertNoError(AssertIdIsNotNull(idNonZero, ""))
		t.AssertError(AssertIdIsNull(idNonZero))
		t.AssertError(AssertEqual(idNonZero, idZero))
		t.AssertError(AssertEqual(idZero, idNonZero))
	}
}

func TestIdEncodeDecode(t1 *testing.T) {
	t := ui.MakeTestContext(t1)

	for _, formatHash := range formatHashes {
		hash := formatHash.Get()

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

	formatHash := FormatHashBlake2b256

	f.Fuzz(
		func(t1 *testing.T, input string) {
			id, repool := formatHash.GetMarklIdForString(input)
			defer repool()

			if input == "" {
				return
			}

			actual := len(id.String()) - len(formatHash.GetMarklFormatId())
			expected := 59

			if actual != expected {
				t1.Errorf("expected %d but got %d", expected, actual)
			}
		},
	)
}
